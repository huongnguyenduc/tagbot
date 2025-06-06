package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCommands(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	cmd := update.Message.Command()

	switch cmd {
	case "start":
		msg := tgbotapi.NewMessage(chatID, startText())
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	case "help":
		msg := tgbotapi.NewMessage(chatID, helpText())
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	case "tagall":
		mentions := getMentions(chatID)
		if mentions == "" {
			mentions = "No members found to mention."
		}
		msg := tgbotapi.NewMessage(chatID, mentions)
		msg.ParseMode = "MarkdownV2"
		bot.Send(msg)

	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Unknown command. Use /help to see available commands."))
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
		mentions := getMentions(chatID)
		if mentions == "" {
			mentions = "No members found to mention."
		}
		msgText := text + "\n" + mentions
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "MarkdownV2"
		bot.Send(msg)
	}
}

func handleChatMemberUpdate(chatMember *tgbotapi.ChatMemberUpdated) {
	chatID := chatMember.Chat.ID
	userID := chatMember.From.ID
	newStatus := chatMember.NewChatMember.Status

	log.Printf("User %d changed status to %s in chat %d", userID, newStatus, chatID)

	switch newStatus {
	case "left", "kicked":
		err := deleteUser(chatID, userID)
		if err != nil {
			log.Printf("Failed to delete user %d from chat %d: %v", userID, chatID, err)
		}
	}
}

func handleSendMessageToChatGroup(update tgbotapi.Update) {
	switch {
	case update.Message.Text != "":
		chatID, message := detectSendToMessage(update.Message.Text)
		if chatID != 0 && message != "" {
			msg := tgbotapi.NewMessage(chatID, message)
			msg.ParseMode = "MarkdownV2"
			bot.Send(msg)
		}
	case update.Message.Photo != nil && update.Message.Caption != "":
		chatID, message := detectSendToMessage(update.Message.Caption)
		if chatID != 0 && message != "" {
			photo := update.Message.Photo[len(update.Message.Photo)-1] // get highest resolution
			msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(photo.FileID))
			msg.Caption = message
			msg.ParseMode = "MarkdownV2"
			bot.Send(msg)
		}
	}
}
