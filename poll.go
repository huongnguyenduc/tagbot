package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Trigger manual poll and cancel today's scheduled poll
func triggerManualPoll(chatID int64) {
	// Get timezone for this chat
	timezone, _, _, err := getGroupSettings(chatID)
	if err != nil {
		log.Println("Failed to get group settings for manual poll, using default timezone:", err)
		timezone = "Asia/Ho_Chi_Minh"
	}
	
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Invalid timezone '%s' for chat %d, using default UTC\n", timezone, chatID)
		loc = time.UTC
	}
	
	// Check if manual poll was already triggered today
	if wasManualPollTriggeredToday(chatID, loc) {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Manual poll was already triggered today!")
		bot.Send(msg)
		return
	}
	
	// Mark manual poll as triggered
	markManualPollTriggered(chatID, loc)
	
	// Send the poll immediately
	sendPoll(chatID)
	
	// Reschedule tasks to skip today's automatic poll
	scheduleGroupTasks(chatID)
	
	// Send confirmation
	msg := tgbotapi.NewMessage(chatID, "✅ Manual poll triggered! Today's automatic poll has been cancelled.")
	bot.Send(msg)
}

func sendPoll(chatID int64) {
	msgText := "🌞 Good morning! Please vote for today's attendance:\n\n"
	// Reset today's votes
	if _, err := db.Exec("DELETE FROM votes WHERE chat_id = $1", chatID); err != nil {
		log.Println("Failed to clear votes:", err)
	}

	// Inline keyboard buttons
	pollMsg := tgbotapi.NewMessage(chatID, msgText)
	pollMsg.ParseMode = "MarkdownV2"
	pollMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Count me in 🟢", "vote_in"),
			tgbotapi.NewInlineKeyboardButtonData("I'm out 🔴", "vote_out"),
			tgbotapi.NewInlineKeyboardButtonData("Maybe 🤔", "vote_maybe"),
		),
	)
	if _, err := bot.Send(pollMsg); err != nil {
		log.Println("Failed to send poll message:", err)
	}
}

func sendReminder(chatID int64) {
	mentions := getMentions(chatID)
	if mentions == "" {
		mentions = "No active players found."
	}
	text := "⏰ Reminder: The deadline to vote is noon today!\n\n" + mentions

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		log.Println("Failed to send reminder message:", err)
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	user := callback.From

	// Save user if not exists
	saveUser(chatID, user)

	var answer string
	switch callback.Data {
	case "vote_in":
		answer = "IN"
	case "vote_out":
		answer = "OUT"
	case "vote_maybe":
		answer = "MAYBE"
	default:
		return
	}
	saveVote(chatID, user.ID, answer)

	// Acknowledge callback
	ack := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("You voted: %s", answer))
	if _, err := bot.Request(ack); err != nil {
		log.Println("Failed to answer callback:", err)
	}

	// Optionally, update original message or send a confirmation message
}
