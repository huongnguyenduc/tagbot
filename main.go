package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	var err error

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Missing TELEGRAM_BOT_TOKEN environment variable")
	}

	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	initDB()

	// Register commands with Telegram client
	commands := []tgbotapi.BotCommand{
		{Command: "help", Description: "Show help message"},
		{Command: "tagall", Description: "Mention all members"},
	}
	_, err = bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Println("Failed to set bot commands:", err)
	}

	// Check if webhook mode is enabled
	useWebhook := os.Getenv("USE_WEBHOOK")
	if useWebhook == "true" || useWebhook == "1" {
		log.Println("Starting in webhook mode...")
		startWebhookServer()
	} else {
		log.Println("Starting in polling mode...")
		startPolling()
	}
}

// startPolling starts the bot in polling mode (original behavior)
func startPolling() {
	// Remove any existing webhook first
	if err := removeWebhook(); err != nil {
		log.Printf("Warning: Failed to remove existing webhook: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Bot started in polling mode. Waiting for updates...")
	for update := range updates {
		processUpdate(update)
	}
}
