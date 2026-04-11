package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

var startTime time.Time

type Config struct {
	Port string `json:"port"`
}

func logRequest(mode, detail string) {
	t := time.Now().Format("15:04:05")
	fmt.Printf("%s - %10s | %s\n", t, mode, detail)
}

func proxyTracker(w http.ResponseWriter, apiURL string, nick string) {
	logRequest("API REQ", nick)
	c := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:149.0) Gecko/20100101 Firefox/149.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Alt-Used", "api.tracker.gg")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Priority", "u=0, i")

	res, err := c.Do(req)
	if err != nil {
		logRequest("Error", err.Error())
		return
	}
	defer res.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, res.Body)
}

func main() {
	startTime = time.Now()

	configFile, _ := os.ReadFile("config.json")
	var config Config
	json.Unmarshal(configFile, &config)
	if config.Port == "" {
		config.Port = "8080"
	}

	v := color.New(color.FgHiMagenta, color.Bold)
	w_color := color.New(color.FgWhite, color.Bold)
	v.Println(figure.NewFigure("Kartos Rank", "slant", true).String())

	http.HandleFunc("/apiprogram/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		
		info := map[string]string{
			"czas_startu": startTime.Format(time.RFC3339),
		}
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
		b, _ := os.ReadFile("src/ranga.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(b)
	})

	http.HandleFunc("/sesja/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := os.ReadFile("src/sesja.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(b)
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

	w_color.Printf("✅ Serwer działa na portcie %s\n", config.Port)
	http.ListenAndServe(":"+config.Port, nil)
}