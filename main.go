package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/dank/rlapi"
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
	Epic     struct {
		RefreshToken string `json:"refreshToken"`
		AccountID    string `json:"accountId"`
		DisplayName  string `json:"displayName"`
	} `json:"epic"`
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
	rlapiClient *rlapi.PsyNetRPC
	rlapiMu     sync.Mutex
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

	if err != nil {
		fmt.Println("[Config] Nie znaleziono config.json. Utworzono plik.")
		config.Port = "8080"
		config.Settings = map[string]string{
			"sesja_winColor":    "#00ff88",
			"sesja_lossColor":   "#ff3366",
			"sesja_streakColor": "#ffffff",
			"sesja_bgColor":     "#141414",
			"sesja_bgOpacity":   "90",
			"sesja_targetNick":  "Kartos'",
			"rlPort":            "49123",
		}
		saveConfig()
		return
	}

	json.Unmarshal(b, &config)

	needsSave := false
	if config.Port == "" {
		config.Port = "8080"
		needsSave = true
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
		needsSave = true
	}

	if needsSave {
		saveConfig()
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
		"epic":        config.Epic,
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

func getRLAPIClient() (*rlapi.PsyNetRPC, error) {
	rlapiMu.Lock()
	defer rlapiMu.Unlock()

	if rlapiClient != nil {
		return rlapiClient, nil
	}

	if config.Epic.RefreshToken == "" {
		return nil, fmt.Errorf("brak refresh tokenu Epic Games")
	}

	egs := rlapi.NewEGS()
	auth, err := egs.AuthenticateWithRefreshToken(config.Epic.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("błąd auth EGS: %v", err)
	}

	if auth.RefreshToken != config.Epic.RefreshToken {
		config.Epic.RefreshToken = auth.RefreshToken
		config.Epic.AccountID = auth.AccountID
		config.Epic.DisplayName = auth.DisplayName
		go saveConfig()
	}

	code, err := egs.GetExchangeCode(auth.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("błąd exchange code: %v", err)
	}

	authToken, err := egs.ExchangeEOSToken(code)
	if err != nil {
		return nil, fmt.Errorf("błąd EOS token: %v", err)
	}

	psyNet := rlapi.NewPsyNet()
	rpc, err := psyNet.AuthPlayer(authToken.AccessToken, authToken.AccountID, auth.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("błąd PsyNet auth: %v", err)
	}

	rlapiClient = rpc
	return rlapiClient, nil
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

		stateMutex.Lock()
		provider := config.Settings["api_provider"]
		stateMutex.Unlock()

		if provider == "tracker" || provider == "" {
			client := &http.Client{Timeout: 10 * time.Second}
			apiURL := "https://api.tracker.gg/api/v2/rocket-league/standard/profile/epic/" + nick
			req, _ := http.NewRequest("GET", apiURL, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "Tracker.gg error", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			io.Copy(w, resp.Body)
			return
		}

		rpc, err := getRLAPIClient()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.Background()

		targetID := nick
		stateMutex.Lock()
		isMe := strings.EqualFold(nick, "me") || strings.EqualFold(nick, "Twoje Konto") || (config.Epic.DisplayName != "" && strings.EqualFold(nick, config.Epic.DisplayName))
		if isMe && config.Epic.AccountID != "" {
			targetID = config.Epic.AccountID
		}
		stateMutex.Unlock()

		playerID := rlapi.NewPlayerID(rlapi.PlatformEpic, targetID)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		skill, err := rpc.GetPlayerSkill(ctx, playerID)
		if err != nil {
			fmt.Printf("RLAPI Error (GetPlayerSkill): %v\n", err)
			http.Error(w, "Błąd pobierania rangi", http.StatusInternalServerError)
			return
		}

		output := map[string]interface{}{
			"platformInfo": map[string]string{"platformUserHandle": nick},
			"segments": []interface{}{
				map[string]interface{}{
					"type":     "overview",
					"metadata": map[string]interface{}{},
					"stats": map[string]interface{}{
						"rewardLevel": map[string]interface{}{"metadata": map[string]string{"rankName": "Bronze"}},
					},
				},
			},
		}

		playlists := map[int]string{
			10: "Ranked Duel 1v1",
			11: "Ranked Doubles 2v2",
			13: "Ranked Standard 3v3",
			27: "Hoops",
			28: "Rumble",
			29: "Dropshot",
			30: "Snow Day",
			34: "Tournament",
		}

		rankNames := []string{
			"Unranked",
			"Bronze I", "Bronze II", "Bronze III",
			"Silver I", "Silver II", "Silver III",
			"Gold I", "Gold II", "Gold III",
			"Platinum I", "Platinum II", "Platinum III",
			"Diamond I", "Diamond II", "Diamond III",
			"Champion I", "Champion II", "Champion III",
			"Grand Champion I", "Grand Champion II", "Grand Champion III",
			"Supersonic Legend",
		}

		for _, s := range skill.Skills {
			name, ok := playlists[s.Playlist]
			if !ok {
				continue
			}

			rankName := "Unknown"
			if s.Tier >= 0 && s.Tier < len(rankNames) {
				rankName = rankNames[s.Tier]
			}

			segment := map[string]interface{}{
				"type":     "playlist",
				"metadata": map[string]string{"name": name},
				"stats": map[string]interface{}{
					"rating": map[string]interface{}{"value": int(s.MMR*20 + 100)},
					"tier": map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":    rankName,
							"iconUrl": fmt.Sprintf("/img/ranks/%d.png", s.Tier),
						},
					},
					"division": map[string]interface{}{
						"metadata": map[string]interface{}{"name": fmt.Sprintf("Division %d", s.Division+1)},
					},
				},
			}
			output["segments"] = append(output["segments"].([]interface{}), segment)
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": output})
	})

	http.HandleFunc("/apisesja/", func(w http.ResponseWriter, r *http.Request) {
		nick := strings.TrimPrefix(r.URL.Path, "/apisesja/")
		if nick == "" {
			return
		}

		rpc, err := getRLAPIClient()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		history, err := rpc.GetMatchHistory(ctx)
		if err != nil {
			fmt.Printf("RLAPI Error (GetMatchHistory): %v\n", err)
			http.Error(w, "Błąd pobierania historii", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	})

	http.HandleFunc("/ranga/", func(w http.ResponseWriter, r *http.Request) {
		stateMutex.Lock()
		rangaCss := config.Settings["ranga_customCss"]
		stateMutex.Unlock()

		b, err := os.ReadFile("src/ranga.html")
		if err != nil {
			http.Error(w, "Nie znaleziono src/ranga.html", http.StatusNotFound)
			return
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

	dashboardFS := http.FileServer(http.Dir("src/dashboard"))
	http.Handle("/dashboard/", http.StripPrefix("/dashboard/", dashboardFS))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard/", http.StatusFound)
			return
		}
	})

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
			"epic":        config.Epic,
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
			"epic":        config.Epic,
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
				case "login_epic":
					if code, ok := action["code"].(string); ok {
						egs := rlapi.NewEGS()
						auth, err := egs.AuthenticateWithCode(code)
						if err == nil {
							config.Epic.RefreshToken = auth.RefreshToken
							config.Epic.AccountID = auth.AccountID
							config.Epic.DisplayName = auth.DisplayName
							rlapiMu.Lock()
							rlapiClient = nil
							rlapiMu.Unlock()
							go saveConfig()
						}
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
