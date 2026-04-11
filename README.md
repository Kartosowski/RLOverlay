# RankRLOverlay

![License](https://img.shields.io/badge/License-MIT-blueviolet?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![HTML5](https://img.shields.io/badge/HTML5-E34F26?style=for-the-badge&logo=html5&logoColor=white)

> Profesjonalny i lekki zestaw overlayów do Rocket League na stream. Śledź swoją aktualną rangę, MMR oraz statystyki bieżącej sesji w czasie rzeczywistym bezpośrednio w OBS.

---

### 🌟 Funkcje
* **Dwa moduły** – Wybierz overlay z samą rangą lub rozbudowany licznik sesji (Win/Loss/MMR Delta).
* **Auto-update** – Statystyki odświeżają się automatycznie co 3 minuty.
* **Pełna kontrola trybów** – Obsługa 1s, 2s oraz 3s bezpośrednio przez prostą zmianę w adresie URL.
* **Modern Design** – Futurystyczny wygląd (font Rajdhani), animacje i przezroczyste tło idealne pod stream.
* **Wydajność** – Backend napisany w Go gwarantuje znikome zużycie zasobów procesora.

---

### 🚀 Szybki start
1. Pobierz najnowszą wersję z sekcji **[Releases](https://github.com/Kartosowski/RankRLOverlay/releases)**.
2. Rozpakuj archiwum i uruchom plik `RankRLOverlay.exe`.
3. Pozostaw okno konsoli otwarte podczas korzystania z overlayów.

---

### 📺 Konfiguracja w OBS
Aby dodać wybrany overlay, dodaj w OBS nowe **Źródło przeglądarki (Browser Source)** i użyj jednego z poniższych formatów linków. 

> **Uwaga:** Zmień `[tryb]` na jeden z trzech: `1s`, `2s` lub `3s`. Zmień `[nick]` na swoją nazwę użytkownika z Epic Games.

#### 1. Aktualna Ranga
Wyświetla ikonę rangi, nazwę dywizji oraz punkty MMR z płynną animacją wejścia.
* **URL:** `http://localhost:8080/ranga/[tryb]/[nick]`
* **Przykład:** `http://localhost:8080/ranga/2s/Kart0s.`

<img width="757" height="356" alt="image" src="https://github.com/user-attachments/assets/212b2864-6f9c-4500-8578-f371182692de" />

#### 2. Statystyki Sesji (Live Tracker)
Wyświetla liczbę zwycięstw, porażek, zmianę MMR w trakcie sesji (Delta) oraz aktualny Streak.
* **URL:** `http://localhost:8080/sesja/[tryb]/[nick]`
* **Przykład:** `http://localhost:8080/sesja/2s/Kart0s.`

<img width="621" height="390" alt="image" src="https://github.com/user-attachments/assets/1019f8dd-b370-48ba-ab31-90dd11226084" />


---

### 🤝 Support i Kontakt
Masz pytania, błędy lub propozycje nowych funkcji?
* **Discord:** [https://discord.gg/wnwCtbe5Ja](https://discord.gg/wnwCtbe5Ja)
* **GitHub:** [https://github.com/Kartosowski/RankRLOverlay](https://github.com/Kartosowski/RankRLOverlay)

---
*Created with ❤️ by Kartosowski*
