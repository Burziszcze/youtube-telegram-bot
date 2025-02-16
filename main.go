package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	TelegramToken string   `json:"telegram_token"`
	ChatID        string   `json:"chat_id"`
	YouTubeAPIKey string   `json:"youtube_api_key"`
	Channels      []string `json:"channels"`
}

type YouTubeResponse struct {
	Items []struct {
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

var (
	config     *Config
	lastVideos map[string]string
)

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	log.Println("📄 Załadowano nową konfigurację.")
	return &cfg, nil
}

func watchConfig(filename string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Błąd tworzenia watcher'a:", err)
	}
	defer watcher.Close()

	err = watcher.Add(filename)
	if err != nil {
		log.Fatal("Nie można obserwować pliku:", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("🔄 Wykryto zmianę w pliku config.json, ponowne wczytywanie...")
				newConfig, err := loadConfig(filename)
				if err != nil {
					log.Println("Błąd odczytu nowej konfiguracji:", err)
					continue
				}
				config = newConfig // Aktualizacja konfiguracji
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Błąd obserwowania pliku:", err)
		}
	}
}

func fetchLatestVideo(apiKey, channelID string) (string, string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?key=%s&channelId=%s&part=snippet,id&order=date&maxResults=1&type=video", apiKey, channelID)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var result YouTubeResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", "", err
	}

	if len(result.Items) > 0 {
		title := result.Items[0].Snippet.Title
		videoID := result.Items[0].ID.VideoID
		return title, videoID, nil
	}

	return "", "", fmt.Errorf("brak nowych filmów")
}

func main() {
	var err error
	config, err = loadConfig("config.json")
	if err != nil {
		log.Fatalf("Nie udało się załadować konfiguracji: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Błąd połączenia z Telegram API: %v", err)
	}

	log.Printf("🤖 Bot zalogowany jako: %s", bot.Self.UserName)

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	lastVideos = make(map[string]string)

	// Uruchomienie obserwowania zmian w pliku config.json
	go watchConfig("config.json")

	for range ticker.C {
		for _, channel := range config.Channels {
			title, videoID, err := fetchLatestVideo(config.YouTubeAPIKey, channel)
			if err != nil {
				log.Println("Błąd pobierania wideo:", err)
				continue
			}

			if lastVideos[channel] != videoID {
				msgText := fmt.Sprintf("🎬 Nowy film: *%s*\n🔗 [Obejrzyj na YouTube](https://www.youtube.com/watch?v=%s)", title, videoID)
				msg := tgbotapi.NewMessageToChannel(config.ChatID, msgText)
				msg.ParseMode = "Markdown"

				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Błąd wysyłania wiadomości:", err)
				} else {
					lastVideos[channel] = videoID
					log.Printf("✅ Wysłano nowe powiadomienie: %s", title)
				}
			}
		}
	}
}
