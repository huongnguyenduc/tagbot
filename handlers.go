package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCommands(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	cmd := update.Message.Command()
	args := update.Message.CommandArguments()

	switch cmd {
	case "help":
		msg := tgbotapi.NewMessage(chatID, helpText())
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	case "settimezone":
		if args == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Usage: /settimezone <timezone>\nExample: /settimezone Asia/Ho_Chi_Minh"))
			return
		}
		resp := setGroupSetting(chatID, "timezone", args)
		bot.Send(tgbotapi.NewMessage(chatID, resp))

	case "setpolltime":
		if args == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Usage: /setpolltime HH:MM\nExample: /setpolltime 10:10"))
			return
		}
		resp := setGroupSetting(chatID, "poll_time", args)
		bot.Send(tgbotapi.NewMessage(chatID, resp))

	case "setremindtime":
		if args == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Usage: /setremindtime HH:MM\nExample: /setremindtime 12:15"))
			return
		}
		resp := setGroupSetting(chatID, "remind_time", args)
		bot.Send(tgbotapi.NewMessage(chatID, resp))

	case "votes":
		counts := listVotes(chatID)
		msg := tgbotapi.NewMessage(chatID, counts)
		bot.Send(msg)

	case "tagall":
		mentions := getMentions(chatID)
		if mentions == "" {
			mentions = "No members found to mention."
		}
		msg := tgbotapi.NewMessage(chatID, mentions)
		msg.ParseMode = "MarkdownV2"
		bot.Send(msg)

	case "poll":
		triggerManualPoll(chatID)

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
