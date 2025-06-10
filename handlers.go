package main

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCommands(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	cmd := update.Message.Command()

	LogInfo("Received command %s from chat %d", cmd, chatID)

	switch cmd {
	case "start":
		msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(startText()))
		msg.ParseMode = "MarkdownV2"
		if _, err := bot.Send(msg); err != nil {
			LogError("Failed to send start message to chat %d: %v", chatID, err)
		}

	case "help":
		msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(helpText()))
		msg.ParseMode = "MarkdownV2"
		if _, err := bot.Send(msg); err != nil {
			LogError("Failed to send help message to chat %d: %v", chatID, err)
		}

	case "all":
		// Get the message text from the command
		message := update.Message.Text
		if message == "" {
			message = "No message provided."
		}

		mentions := getMentions(chatID)
		if mentions == "" {
			mentions = "No members found to mention."
		}
		msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(message)+"\n"+mentions)
		msg.ParseMode = "MarkdownV2"
		if _, err := bot.Send(msg); err != nil {
			LogError("Failed to send all message to chat %d: %v", chatID, err)
		}
	}
}

func handleAtAllMention(update tgbotapi.Update) {
	// Ignore messages not from users (e.g., from the bot itself)
	if update.Message == nil || update.Message.From.IsBot {
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if strings.Contains(strings.ToLower(text), "@all") {
		LogInfo("Received @all mention in chat %d from user %d", chatID, update.Message.From.ID)
		mentions := getMentions(chatID)
		if mentions == "" {
			mentions = "No members found to mention."
		}
		msgText := escapeMarkdownV2(text) + "\n" + mentions
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "MarkdownV2"
		if _, err := bot.Send(msg); err != nil {
			LogError("Failed to send @all mention message to chat %d: %v", chatID, err)
		}
	}
}

func handleChatMemberUpdate(chatMember *tgbotapi.ChatMemberUpdated) {
	chatID := chatMember.Chat.ID
	userID := chatMember.From.ID
	newStatus := chatMember.NewChatMember.Status

	LogInfo("User %d changed status to %s in chat %d", userID, newStatus, chatID)

	switch newStatus {
	case "left", "kicked":
		err := deleteUser(chatID, userID)
		if err != nil {
			LogError("Failed to delete user %d from chat %d: %v", userID, chatID, err)
		}
	}
}

func handleSendMessageToChatGroup(update tgbotapi.Update) {
	switch {
	case update.Message.Text != "":
		chatID, message := detectSendToMessage(update.Message.Text)
		if chatID != 0 && message != "" {
			msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(message))
			msg.ParseMode = "MarkdownV2"
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to send message to chat %d: %v", chatID, err)
			}
		}
	case update.Message.Photo != nil && update.Message.Caption != "":
		chatID, message := detectSendToMessage(update.Message.Caption)
		if chatID != 0 {
			photo := update.Message.Photo[len(update.Message.Photo)-1] // get highest resolution
			msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(photo.FileID))
			if message != "" {
				msg.Caption = escapeMarkdownV2(message)
				msg.ParseMode = "MarkdownV2"
			}
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to send message to chat %d: %v", chatID, err)
			}
		}
	case update.Message.Document != nil:
		chatID, message := detectSendToMessage(update.Message.Document.FileName)
		if chatID != 0 {
			msg := tgbotapi.NewDocument(chatID, tgbotapi.FileID(update.Message.Document.FileID))
			if message != "" {
				msg.Caption = escapeMarkdownV2(message)
				msg.ParseMode = "MarkdownV2"
			}
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to send document to chat %d: %v", chatID, err)
			}
		}
	}
}
