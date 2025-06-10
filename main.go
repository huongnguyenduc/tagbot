package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Initialize logger
	initLogger()
	LogInfo("Starting TagBot...")

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		LogFatal("Missing TELEGRAM_BOT_TOKEN environment variable")
	}

	var err error
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		LogFatal("Failed to create bot: %v", err)
	}

	bot.Debug = false
	LogInfo("Authorized on account %s", bot.Self.UserName)

	initDB()

	// Register commands with Telegram client
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Show welcome message"},
		{Command: "help", Description: "Show help message"},
		{Command: "all", Description: "Mention all members"},
	}
	_, err = bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		LogError("Failed to set bot commands: %v", err)
	}

	// Check if webhook mode is enabled
	useWebhook := os.Getenv("USE_WEBHOOK")
	if useWebhook == "true" || useWebhook == "1" {
		LogInfo("Starting in webhook mode...")
		startWebhookServer()
	} else {
		LogInfo("Starting in polling mode...")
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
