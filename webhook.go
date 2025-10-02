package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebhookHandler handles incoming webhook requests from Telegram
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		LogError("Received non-POST request: %s", r.Method)
		return
	}

	// Parse the update from the request body
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		LogError("Failed to decode webhook update: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process the update
	processUpdate(update)

	// Respond with 200 OK
	w.WriteHeader(http.StatusOK)
}

// processUpdate handles a single update (shared between polling and webhook)
func processUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID

		// Save user to DB on any message
		if update.Message.From != nil {
			saveUser(chatID, update.Message.From)
		}

		// Handle commands
		if update.Message.IsCommand() {
			handleCommands(update)
			return
		}

		// Handle @all mentions
		handleAtAllMention(update)

		// Handle chat member updates
		if update.ChatMember != nil {
			handleChatMemberUpdate(update.ChatMember)
		}

		// Handle send message to chat group
		handleSendMessageToChatGroup(update)

		// Handle forward message to special chat
		handleForwardMessageToSpecialChat(update)
	}
}

// setupWebhook configures the webhook with Telegram
func setupWebhook(webhookURL string) error {
	LogInfo("Setting up webhook at: %s", webhookURL)

	webhook, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	_, err = bot.Request(webhook)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	LogInfo("Webhook set successfully")
	return nil
}

// removeWebhook removes the webhook (useful for switching back to polling)
func removeWebhook() error {
	LogInfo("Removing webhook...")

	_, err := bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		return fmt.Errorf("failed to remove webhook: %w", err)
	}

	LogInfo("Webhook removed successfully")
	return nil
}

// startWebhookServer starts the HTTP server for webhook handling
func startWebhookServer() {
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		LogFatal("WEBHOOK_URL environment variable is required for webhook mode")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Set up the webhook with Telegram
	if err := setupWebhook(webhookURL); err != nil {
		LogFatal("Failed to setup webhook: %v", err)
	}

	// Set up HTTP server
	http.HandleFunc("/webhook", webhookHandler)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	LogInfo("Starting webhook server on port %s", port)
	LogInfo("Webhook endpoint: /webhook")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		LogFatal("Failed to start webhook server: %v", err)
	}
}
