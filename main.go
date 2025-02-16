package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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
	err = yaml.Unmarshal(body, &result)
	if err != nil {
		return "", "", err
	}

	if len(result.Items) > 0 {
		title := result.Items[0].Snippet.Title
		videoID := result.Items[0].ID.VideoID
		return title, videoID, nil
	}

	return "", "", fmt.Errorf("no new videos found")
}

func loadLastVideos(filename string) map[string]string {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Println("No previous video history found, creating new one.")
		return make(map[string]string)
	}

	var videos map[string]string
	err = yaml.Unmarshal(data, &videos)
	if err != nil {
		log.Println("Error loading last videos file:", err)
		return make(map[string]string)
	}

	return videos
}

func saveLastVideos(filename string, data map[string]string) {
	jsonData, err := yaml.Marshal(data)
	if err != nil {
		log.Println("Error saving last videos:", err)
		return
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		log.Println("Error writing last videos file:", err)
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

	lastVideos = loadLastVideos(lastVideosFile)

	go watchConfig(configFile)

	for range ticker.C {
		for _, channel := range config.Channels {
			title, videoID, err := fetchLatestVideo(config.YouTubeAPIKey, channel)
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
					saveLastVideos(lastVideosFile, lastVideos)
					log.Printf("✅ Sent new video notification: %s", title)
				}
			}
		}
	}
}
