package main

import (
	"fmt"
	"log"
	"os"
	"time"

	youtube "github.com/Burziszcze/youtube-telegram-bot/youtube" // Import nowego modułu
	"github.com/fsnotify/fsnotify"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/go-yaml/yaml"
)

type Config struct {
	TelegramToken string   `yaml:"telegram_token"`
	ChatID        string   `yaml:"chat_id"`
	YouTubeAPIKey string   `yaml:"youtube_api_key"`
	Channels      []string `yaml:"channels"`
}

var (
	config     *Config
	lastVideos map[string]string
)

const (
	configFile     = "config.yml"
	lastVideosFile = "last_videos.json"
)

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	log.Println("📄 Loaded new configuration.")
	return &cfg, nil
}

func watchConfig(filename string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	err = watcher.Add(filename)
	if err != nil {
		log.Fatal("Cannot watch file:", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("🔄 Detected changes in config.yml, reloading...")
				newConfig, err := loadConfig(filename)
				if err != nil {
					log.Println("Error reading new configuration:", err)
					continue
				}
				config = newConfig
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func main() {
	var err error
	config, err = loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Error connecting to Telegram API: %v", err)
	}

	log.Printf("🤖 Bot logged in as: %s", bot.Self.UserName)

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	lastVideos = youtube.LoadLastVideos(lastVideosFile) // Użycie funkcji z nowego modułu

	go watchConfig(configFile)

	for range ticker.C {
		for _, channel := range config.Channels {
			title, videoID, err := youtube.FetchLatestVideo(config.YouTubeAPIKey, channel) // Użycie funkcji z nowego modułu
			if err != nil {
				log.Println("Error fetching video:", err)
				continue
			}

			if lastVideos[channel] != videoID {
				msgText := fmt.Sprintf("🎬 New video: *%s*\n🔗 [Watch on YouTube](https://www.youtube.com/watch?v=%s)", title, videoID)
				msg := tgbotapi.NewMessageToChannel(config.ChatID, msgText)
				msg.ParseMode = "Markdown"

				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Error sending message:", err)
				} else {
					lastVideos[channel] = videoID
					youtube.SaveLastVideos(lastVideosFile, lastVideos) // Użycie funkcji z nowego modułu
					log.Printf("✅ Sent new video notification: %s", title)
				}
			}
		}
	}
}
