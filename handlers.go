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
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if strings.Contains(strings.ToLower(text), "@all") {
		mentions := getMentions(chatID)
		trimmedText := removeAtAll(text)
		if mentions == "" {
			mentions = "No members found to mention."
		}
		msgText := trimmedText + "\n\n" + mentions
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
