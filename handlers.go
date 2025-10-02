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

// Hidden command to send message to chat group by @sendto <chat_id> <message>
func handleSendMessageToChatGroup(update tgbotapi.Update) {
	switch {
	// Handle text messages
	case update.Message.Text != "":
		chatID, message := detectSendToMessage(update.Message.Text)
		if chatID != 0 && message != "" {
			msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2(message))
			msg.ParseMode = "MarkdownV2"
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to send message to chat %d: %v", chatID, err)
			}
		}
	// Handle photo messages
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
	// Handle document messages
	case update.Message.Document != nil && update.Message.Caption != "":
		chatID, message := detectSendToMessage(update.Message.Caption)
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
	// Handle poll messages
	case update.Message.Poll != nil && !update.Message.Poll.IsClosed && update.Message.Poll.Question != "":
		chatID, question := detectSendToMessage(update.Message.Poll.Question)
		if chatID != 0 && question != "" {
			options := make([]string, len(update.Message.Poll.Options))
			for i, opt := range update.Message.Poll.Options {
				options[i] = opt.Text
			}
			msg := tgbotapi.NewPoll(chatID, question, options...)
			msg.IsAnonymous = update.Message.Poll.IsAnonymous
			msg.AllowsMultipleAnswers = update.Message.Poll.AllowsMultipleAnswers
			msg.Type = update.Message.Poll.Type
			msg.Explanation = update.Message.Poll.Explanation
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to send poll to chat %d: %v", chatID, err)
			}
		}
	}
}

func handleForwardMessageToSpecialChat(update tgbotapi.Update) {
	if len(specialChatIDs) == 0 {
		return // No special chats configured
	}

	for _, specialChatID := range specialChatIDs {

		// Forward the message based on its type
		switch {
		case update.Message.Text != "":
			// Forward text message
			msg := tgbotapi.NewMessage(specialChatID, update.Message.Text)
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward text message to chat %d: %v", specialChatID, err)
			}

		case update.Message.Photo != nil:
			// Forward photo message
			photo := update.Message.Photo[len(update.Message.Photo)-1] // get highest resolution
			msg := tgbotapi.NewPhoto(specialChatID, tgbotapi.FileID(photo.FileID))
			if update.Message.Caption != "" {
				msg.Caption = update.Message.Caption
			}
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward photo to chat %d: %v", specialChatID, err)
			}

		case update.Message.Document != nil:
			// Forward document message
			msg := tgbotapi.NewDocument(specialChatID, tgbotapi.FileID(update.Message.Document.FileID))
			if update.Message.Caption != "" {
				msg.Caption = update.Message.Caption
			}
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward document to chat %d: %v", specialChatID, err)
			}

		case update.Message.Video != nil:
			// Forward video message
			msg := tgbotapi.NewVideo(specialChatID, tgbotapi.FileID(update.Message.Video.FileID))
			if update.Message.Caption != "" {
				msg.Caption = update.Message.Caption
			}
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward video to chat %d: %v", specialChatID, err)
			}

		case update.Message.Audio != nil:
			// Forward audio message
			msg := tgbotapi.NewAudio(specialChatID, tgbotapi.FileID(update.Message.Audio.FileID))
			if update.Message.Caption != "" {
				msg.Caption = update.Message.Caption
			}
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward audio to chat %d: %v", specialChatID, err)
			}

		case update.Message.Voice != nil:
			// Forward voice message
			msg := tgbotapi.NewVoice(specialChatID, tgbotapi.FileID(update.Message.Voice.FileID))
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward voice to chat %d: %v", specialChatID, err)
			}

		case update.Message.Sticker != nil:
			// Forward sticker message
			msg := tgbotapi.NewSticker(specialChatID, tgbotapi.FileID(update.Message.Sticker.FileID))
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward sticker to chat %d: %v", specialChatID, err)
			}

		case update.Message.Poll != nil:
			// Forward poll message
			options := make([]string, len(update.Message.Poll.Options))
			for i, opt := range update.Message.Poll.Options {
				options[i] = opt.Text
			}
			msg := tgbotapi.NewPoll(specialChatID, update.Message.Poll.Question, options...)
			msg.IsAnonymous = update.Message.Poll.IsAnonymous
			msg.AllowsMultipleAnswers = update.Message.Poll.AllowsMultipleAnswers
			msg.Type = update.Message.Poll.Type
			msg.Explanation = update.Message.Poll.Explanation
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward poll to chat %d: %v", specialChatID, err)
			}

		default:
			// Forward as generic message if type is not supported
			msg := tgbotapi.NewMessage(specialChatID, "Unsupported message type forwarded")
			if _, err := bot.Send(msg); err != nil {
				LogError("Failed to forward unsupported message to chat %d: %v", specialChatID, err)
			}
		}
	}
}
