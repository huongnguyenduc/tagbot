package main

import (
	"fmt"
	"log"
	"strings"
)

// Escape MarkdownV2 special chars for Telegram
func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

// Remove @all mention from text to avoid double mention issues
func removeAtAll(text string) string {
	words := strings.Fields(text)
	var filtered []string
	for _, w := range words {
		if strings.ToLower(w) != "@all" {
			filtered = append(filtered, w)
		}
	}
	return strings.Join(filtered, " ")
}

func getMentions(chatID int64) string {
	query := `SELECT user_id, first_name, last_name, username FROM members WHERE chat_id = $1`
	rows, err := db.Query(query, chatID)
	if err != nil {
		log.Println("Query failed:", err)
		return ""
	}
	defer rows.Close()

	var mentions []string
	for rows.Next() {
		var userID int64
		var firstName, lastName, username string
		if err := rows.Scan(&userID, &firstName, &lastName, &username); err != nil {
			continue
		}
		if username != "" {
			mentions = append(mentions, "@"+escapeMarkdownV2(username))
		} else {
			name := escapeMarkdownV2(firstName)
			if lastName != "" {
				name += " " + escapeMarkdownV2(lastName)
			}
			mentions = append(mentions, fmt.Sprintf("[%s](tg://user?id=%d)", name, userID))
		}
	}
	return strings.Join(mentions, " ")
}

func helpText() string {
	return "🤖 *Werewolf Bot Help*\n\n" +
		"Use /help to see this message.\n\n" +
		"*Commands:*\n" +
		"/help - Show this help message\n" +
		"/poll - Trigger daily poll manually (cancels today's automatic poll)\n" +
		"/settimezone <timezone> - Set group timezone (e.g. Asia/Ho_Chi_Minh)\n" +
		"/setpolltime HH:MM - Set poll time (24h format)\n" +
		"/setremindtime HH:MM - Set reminder time (24h format)\n" +
		"/votes - Show current vote counts\n" +
		"/tagall - Mention all members\n\n" +
		"To vote, use the inline buttons in the poll message.\n" +
		"@all or /tagall will mention all players in the group."
}

func setGroupSetting(chatID int64, field, value string) string {
	err := updateGroupSetting(chatID, field, value)
	if err != nil {
		return fmt.Sprintf("Failed to update %s: %v", field, err)
	}
	// Reschedule after updating
	scheduleGroupTasks(chatID)
	return fmt.Sprintf("Updated %s to %s", field, value)
}
