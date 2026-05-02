# RL OVERLAY
<img width="1919" height="1079" alt="image" src="https://github.com/user-attachments/assets/a93d7ea0-be76-42d1-b979-8919f2e0c42e" />

---

## JAK POBRAĆ?
1. Przejdź do zakładki [**Releases**](https://github.com/Kartosowski/RLOverlay/releases) i pobierz najnowszą wersję.
2. Rozpakuj pobraną paczkę na swoim dysku.
3. Uruchom plik **`RLOverlay.exe`**.
4. Domyślny port aplikacji to **8080**.

---

## DASHBOARD - Zarządzanie
Pod adresem `http://localhost:8080/dashboard/` możesz:
- Zmieniać kolory nakładek.
- Resetować statystyki sesji.
- Zmieniać ustawienia przezroczystości.
- Całość jest bardzo łatwa w obsłudze.

---

## SESJE - Jak skonfigurować?
Statystyki sesji działają w **Real-Time**, co oznacza, że serwer automatycznie wykrywa wynik meczu natychmiast po jego zakończeniu.

1. W folderze:  
   `C:\Users\[Twoja Nazwa]\Documents\My Games\Rocket League\TAGame\Config`  
   edytuj plik **`TAStatsAPI.ini`**.
2. Zmień wartość `PacketSendRate=0` na **`PacketSendRate=1`**.
3. Upewnij się, że port jest ustawiony na **`49123`** (możesz go zmienić na dowolny, ale pamiętaj, aby zaktualizować go również w Dashboardzie!).
4. Na Dashboardzie (`http://localhost:8080/dashboard/`) ustaw swój **nick**, którego aktualnie używasz w grze.
5. Dodaj nowe **Źródło Przeglądarki** w OBS i wklej link:  
   `http://localhost:8080/sesja/`

<img width="504" height="178" alt="image" src="https://github.com/user-attachments/assets/5a968d1a-5f9c-4a4c-8065-18fff2dfbd16" />

---

## RANGA - Jak skonfigurować?
Nakładka rangi pobiera dane z Twojego profilu i wyświetla aktualną rangę oraz MMR.

1. Dodaj **Źródło Przeglądarki** w OBS.
2. Wklej link w formacie:  
   `http://localhost:8080/ranga/[tryb]/[Twój_Nick]`  
   *(Dostępne tryby: 1s, 2s, 3s)*.
3. Przykładowy link dla trybu 2v2:  
   `http://localhost:8080/ranga/2s/Kartos`
   
<img width="1028" height="298" alt="image" src="https://github.com/user-attachments/assets/ca387ab9-ce11-40bb-9496-cc56ff33fc2b" />

---

### Pomoc / Support
Masz problem z konfiguracją? Dołącz do serwera Discord:  
[**https://discord.gg/xRJvJCzhWp**](https://discord.gg/xRJvJCzhWp)

---

### Wykonawca: Kartos
[GitHub Projektu](https://github.com/Kartosowski/RLOverlay)
