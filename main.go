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

type Config struct {
	Port string `json:"port"`
}

func logRequest(mode, detail string) {
    t := time.Now().Format("15:04:05")
    fmt.Printf("%s - %10s | %s\n", t, mode, detail)
}

func main() {
	configFile, _ := os.ReadFile("config.json")
	var config Config
	json.Unmarshal(configFile, &config)
	if config.Port == "" { config.Port = "8080" }

	v := color.New(color.FgHiMagenta, color.Bold)
	w := color.New(color.FgWhite, color.Bold)

	myFigure := figure.NewFigure("Kartos Rank", "slant", true)
	v.Println(myFigure.String())
	
	w.Println(" Github:  https://github.com/Kartosowski/RankRL")
	w.Println(" Discord: https://discord.gg/wnwCtbe5Ja\n")

	
	fmt.Printf(" 1v1: http://localhost:%s/1s/NICK\n", config.Port)
	fmt.Printf(" 2v2: http://localhost:%s/2s/NICK\n", config.Port)
	fmt.Printf(" 3v3: http://localhost:%s/3s/NICK\n\n (W nicku musisz dać nazwę z Epic Games.)\n\n", config.Port)

	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(r.URL.Path, "/")
		if len(p) < 3 { return }
		nick := p[2]
		
		logRequest("Request API", nick)

		c := &http.Client{Timeout: 10 * time.Second}
		req, _ := http.NewRequest("GET", "https://api.tracker.gg/api/v2/rocket-league/standard/profile/epic/"+nick, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		
		res, err := c.Do(req)
		if err != nil {
			logRequest("Error", err.Error())
			return
		}
		defer res.Body.Close()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, res.Body)
	})

	http.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		f := strings.TrimPrefix(r.URL.Path, "/img/")
		b, err := os.ReadFile("img/" + f)
		if err != nil { return }
		w.Header().Set("Content-Type", "image/png")
		w.Write(b)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/img/") { return }
		b, err := os.ReadFile("src/index.html")
		if err != nil {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(b)
	})

	w.Printf("✅ Serwer działa na portcie %s\n", config.Port)
	http.ListenAndServe(":"+config.Port, nil)
}