package modules

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Bot struct {
	API     *tgbotapi.BotAPI
	YouTube *youtube.Service
}

func NewBot(telegramToken, youtubeAPIKey string) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %v", err)
	}

	youtubeService, err := youtube.NewService(nil, option.WithAPIKey(youtubeAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %v", err)
	}

	return &Bot{
		API:     botAPI,
		YouTube: youtubeService,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.API.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("failed to get updates: %v", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		}
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStartCommand(message)
	case "search":
		b.handleSearchCommand(message)
	default:
		b.handleUnknownCommand(message)
	}
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome to the YouTube Telegram Bot!")
	b.API.Send(msg)
}

func (b *Bot) handleSearchCommand(message *tgbotapi.Message) {
	query := message.CommandArguments()
	if query == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a search query.")
		b.API.Send(msg)
		return
	}

	searchCall := b.YouTube.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(5)
	response, err := searchCall.Do()
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Failed to search YouTube: %v", err))
		b.API.Send(msg)
		return
	}

	var results []string
	for _, item := range response.Items {
		results = append(results, fmt.Sprintf("Title: %s\nURL: https://www.youtube.com/watch?v=%s", item.Snippet.Title, item.Id.VideoId))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, strings.Join(results, "\n\n"))
	b.API.Send(msg)
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Please use /start or /search.")
	b.API.Send(msg)
}
