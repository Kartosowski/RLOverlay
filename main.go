package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var startTime time.Time

type Config struct {
	Port     string            `json:"port"`
	Settings map[string]string `json:"settings"`
}

type SessionState struct {
	Wins             int      `json:"wins"`
	Losses           int      `json:"losses"`
	Streak           int      `json:"streak"`
	ProcessedMatches []string `json:"processedMatches"`
}

var (
	stateMutex  sync.Mutex
	state       SessionState
	config      Config
	clients     = make(map[*websocket.Conn]bool)
	dashClients = make(map[*websocket.Conn]bool)
	rlConnected bool
)

func loadState() {
	b, err := os.ReadFile("sesja.json")
	if err == nil {
		json.Unmarshal(b, &state)
	}
	if state.ProcessedMatches == nil {
		state.ProcessedMatches = []string{}
	}
}

func loadConfig() {
	b, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(b, &config)
	}
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Settings == nil {
		config.Settings = map[string]string{
			"sesja_winColor":    "#00ff88",
			"sesja_lossColor":   "#ff3366",
			"sesja_streakColor": "#ffffff",
			"sesja_bgColor":     "#141414",
			"sesja_bgOpacity":   "90",
			"sesja_targetNick":  "Kartos'",
			"rlPort":            "49123",
		}
	}
}

func saveConfig() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	b, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("config.json", b, 0644)
}

func saveState() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	b, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile("sesja.json", b, 0644)
}

func broadcastState() {
	stateMutex.Lock()
	msg := map[string]interface{}{
		"wins":        state.Wins,
		"losses":      state.Losses,
		"streak":      state.Streak,
		"settings":    config.Settings,
		"rlConnected": rlConnected,
	}
	stateMutex.Unlock()

	for client := range clients {
		client.WriteJSON(msg)
	}
	for client := range dashClients {
		client.WriteJSON(msg)
	}
}

func hasProcessed(guid string) bool {
	for _, m := range state.ProcessedMatches {
		if m == guid {
			return true
		}
	}
	return false
}

func processGameData(dataStr string) {
	var gd struct {
		MatchGuid string `json:"MatchGuid"`
		Game      struct {
			BHasWinner bool   `json:"bHasWinner"`
			Winner     string `json:"Winner"`
			Teams      []struct {
				Name    string `json:"Name"`
				TeamNum int    `json:"TeamNum"`
			} `json:"Teams"`
		} `json:"Game"`
		Players []struct {
			Name    string `json:"Name"`
			TeamNum int    `json:"TeamNum"`
		} `json:"Players"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(dataStr)), &gd); err != nil {
		return
	}

	stateMutex.Lock()
	targetNick := config.Settings["sesja_targetNick"]
	stateMutex.Unlock()
	if targetNick == "" {
		targetNick = "Kartos'"
	}

	myTeamNum := -1
	for _, p := range gd.Players {
		if p.Name == targetNick {
			myTeamNum = p.TeamNum
		}
	}

	gd.MatchGuid = strings.ToUpper(strings.TrimSpace(gd.MatchGuid))
	if myTeamNum != -1 && gd.Game.BHasWinner && gd.MatchGuid != "" && gd.MatchGuid != "00000000000000000000000000000000" && gd.MatchGuid != "00000000-0000-0000-0000-000000000000" {
		stateMutex.Lock()
		if !hasProcessed(gd.MatchGuid) {
			state.ProcessedMatches = append(state.ProcessedMatches, gd.MatchGuid)

			winnerTeamNum := -1
			for _, t := range gd.Game.Teams {
				if t.Name == gd.Game.Winner {
					winnerTeamNum = t.TeamNum
					break
				}
			}

			if winnerTeamNum == myTeamNum {
				state.Wins++
				if state.Streak > 0 {
					state.Streak++
				} else {
					state.Streak = 1
				}
			} else {
				state.Losses++
				if state.Streak < 0 {
					state.Streak--
				} else {
					state.Streak = -1
				}
			}
			stateMutex.Unlock()
			saveState()
			broadcastState()
		} else {
			stateMutex.Unlock()
		}
	}
}

func startRLTCPClient() {
	for {
		stateMutex.Lock()
		rlPort := config.Settings["rlPort"]
		stateMutex.Unlock()

		if rlPort == "" {
			rlPort = "49123"
		}

		conn, err := net.Dial("tcp", "127.0.0.1:"+rlPort)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		stateMutex.Lock()
		rlConnected = true
		stateMutex.Unlock()
		broadcastState()

		dec := json.NewDecoder(conn)
		for {
			var pkt struct {
				Event string `json:"Event"`
				Data  string `json:"Data"`
			}
			if err := dec.Decode(&pkt); err != nil {
				break
			}

			if pkt.Event == "UpdateState" {
				processGameData(pkt.Data)
			}
		}

		conn.Close()
		stateMutex.Lock()
		rlConnected = false
		stateMutex.Unlock()
		broadcastState()

		fmt.Println("[RL] ❌ Utracono połączenie — retry za 5s...")
		time.Sleep(5 * time.Second)
	}
}

func proxyTracker(w http.ResponseWriter, apiURL string, nick string) {
	c := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Storage-Access", "active")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	res, err := c.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, res.Body)
}

func main() {
	startTime = time.Now()
	loadConfig()
	loadState()

	go startRLTCPClient()

	cyan := color.New(color.FgCyan)
	white := color.New(color.FgWhite)
	whiteBold := color.New(color.FgWhite, color.Bold)

	whiteBold.Println("RL Overlay")
	white.Println("Wykonawca: Kartos")
	white.Printf("Aplikacja dziala na porcie: %s\n", config.Port)
	white.Print("Dashboard: ")
	cyan.Printf("http://localhost:%s/dashboard/\n", config.Port)
	white.Print("GitHub:    ")
	cyan.Println("https://github.com/Kartosowski/RLOverlay")
	fmt.Println()

	http.HandleFunc("/apiprogram/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		info := map[string]string{"czas_startu": startTime.Format(time.RFC3339)}
		json.NewEncoder(w).Encode(info)
	})

	http.HandleFunc("/apiranga/", func(w http.ResponseWriter, r *http.Request) {
		nick := strings.TrimPrefix(r.URL.Path, "/apiranga/")
		if nick == "" {
			return
		}
		apiURL := "https://api.tracker.gg/api/v2/rocket-league/standard/profile/epic/" + nick
		proxyTracker(w, apiURL, nick)
	})

	http.HandleFunc("/apisesja/", func(w http.ResponseWriter, r *http.Request) {
		nick := strings.TrimPrefix(r.URL.Path, "/apisesja/")
		if nick == "" {
			return
		}
		apiURL := "https://api.tracker.gg/api/v2/rocket-league/standard/profile/epic/" + nick + "/sessions/"
		proxyTracker(w, apiURL, nick)
	})

	http.HandleFunc("/ranga/", func(w http.ResponseWriter, r *http.Request) {
		stateMutex.Lock()
		theme := config.Settings["ranga_theme"]
		rangaCss := config.Settings["ranga_customCss"]
		stateMutex.Unlock()

		if theme == "" {
			theme = "default"
		}

		filePath := "src/ranga.html"
		if theme != "default" {
			filePath = filepath.Join("custom_overlay", "ranga", theme, "ranga.html")
		}

		b, err := os.ReadFile(filePath)
		if err != nil {
			b, _ = os.ReadFile("src/ranga.html")
		}

		html := string(b)
		if rangaCss != "" {
			if strings.Contains(html, "</head>") {
				html = strings.Replace(html, "</head>", "<style id=\"custom-css\">"+rangaCss+"</style>\n</head>", 1)
			}
		}

		wsScript := `
		<script>
			function connectDashboardReload() {
				const ws = new WebSocket('ws://' + window.location.host + '/ws/sesja');
				ws.onmessage = (event) => {
					const data = JSON.parse(event.data);
					if (data.settings && data.settings.ranga_force_reload === 'true') {
						window.location.reload();
					}
					if (data.settings && data.settings.ranga_customCss !== undefined) {
						let el = document.getElementById('custom-css');
						if (!el) {
							el = document.createElement('style');
							el.id = 'custom-css';
							document.head.appendChild(el);
						}
						el.innerHTML = data.settings.ranga_customCss;
					}
				};
				ws.onclose = () => setTimeout(connectDashboardReload, 2000);
			}
			connectDashboardReload();
		</script>
		`
		html = strings.Replace(html, "</body>", wsScript+"</body>", 1)

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	http.HandleFunc("/sesja/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := os.ReadFile("src/sesja.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(b)
	})

	http.HandleFunc("/api/themes", func(w http.ResponseWriter, r *http.Request) {
		themes := []string{"default"}
		dirs, err := os.ReadDir(filepath.Join("custom_overlay", "ranga"))
		if err == nil {
			for _, d := range dirs {
				if d.IsDir() {
					themes = append(themes, d.Name())
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(themes)
	})

	dashboardFS := http.FileServer(http.Dir("src/dashboard"))
	http.Handle("/dashboard/", http.StripPrefix("/dashboard/", dashboardFS))

	http.HandleFunc("/ws/sesja", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		stateMutex.Lock()
		clients[conn] = true
		msg := map[string]interface{}{
			"wins":        state.Wins,
			"losses":      state.Losses,
			"streak":      state.Streak,
			"settings":    config.Settings,
			"rlConnected": rlConnected,
		}
		conn.WriteJSON(msg)
		stateMutex.Unlock()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}

		stateMutex.Lock()
		delete(clients, conn)
		stateMutex.Unlock()
	})

	http.HandleFunc("/ws/dashboard", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		stateMutex.Lock()
		dashClients[conn] = true
		msg := map[string]interface{}{
			"wins":        state.Wins,
			"losses":      state.Losses,
			"streak":      state.Streak,
			"settings":    config.Settings,
			"rlConnected": rlConnected,
		}
		conn.WriteJSON(msg)
		stateMutex.Unlock()

		for {
			var action map[string]interface{}
			if err := conn.ReadJSON(&action); err != nil {
				break
			}

			if act, ok := action["action"].(string); ok {
				stateMutex.Lock()
				switch act {
				case "add_win":
					state.Wins++
					if state.Streak > 0 {
						state.Streak++
					} else {
						state.Streak = 1
					}
				case "sub_win":
					if state.Wins > 0 {
						state.Wins--
					}
				case "add_loss":
					state.Losses++
					if state.Streak < 0 {
						state.Streak--
					} else {
						state.Streak = -1
					}
				case "sub_loss":
					if state.Losses > 0 {
						state.Losses--
					}
				case "reset":
					state.Wins = 0
					state.Losses = 0
					state.Streak = 0
					state.ProcessedMatches = []string{}
				case "update_settings":
					if settingsMap, ok := action["settings"].(map[string]interface{}); ok {
						for k, v := range settingsMap {
							switch val := v.(type) {
							case string:
								config.Settings[k] = val
							case bool:
								if val {
									config.Settings[k] = "true"
								} else {
									config.Settings[k] = "false"
								}
							}
						}
						go saveConfig()
					}
				}
				stateMutex.Unlock()
				saveState()
				broadcastState()
			}
		}

		stateMutex.Lock()
		delete(dashClients, conn)
		stateMutex.Unlock()
	})

	http.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		f := strings.TrimPrefix(r.URL.Path, "/img/")
		b, err := os.ReadFile("img/" + f)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(b)
	})

	http.ListenAndServe(":"+config.Port, nil)
}
