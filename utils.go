package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

func startText() string {
	return "🌕 *Awooo! I am the Alpha Wolf of @werewolf_u2u_bot!*\n\n" +
		"As the Alpha of this pack, I'll help gather all the wolves for our nightly hunts. Here's how to summon the pack:\n\n" +
		"*Pack Commands:*\n" +
		"• /start - Hear the Alpha's howl\n" +
		"• /help - Learn the ways of the pack\n" +
		"• /all - Summon all wolves to the hunt\n\n" +
		"*Pack Features:*\n" +
		"• Type `@all` in any message to call the pack\n" +
		"• I track all wolves in our territory\n" +
		"• Wolves who leave are removed from the pack\n\n" +
		"*Note:* To summon the pack, I need to be an Alpha in the group. Grant me the necessary permissions to lead the hunt.\n\n"
}

func helpText() string {
	return "🌕 *Pack Commands Guide*\n\n" +
		"*How to Summon the Pack:*\n" +
		"• Use /all to call all wolves to the hunt\n" +
		"• Type @all in any message to gather the pack\n\n" +
		"*Alpha's Notes:*\n" +
		"• I track all wolves in our territory\n" +
		"• Wolves who leave are removed from the pack\n" +
		"• I need to be an Alpha to summon the pack\n\n" +
		"*Need Help?*\n" +
		"Use /start to hear the Alpha's howl again.\n\n"
}

// Detect @sendto <chat_id> <message> pattern & return the chat_id and message
func detectSendToMessage(text string) (int64, string) {
	if strings.HasPrefix(text, "@sendto") {
		parts := strings.Split(text, " ")
		if len(parts) >= 3 {
			chatIDStr := parts[1]
			chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
			if err != nil {
				log.Printf("Invalid chatID: %s, error: %v", chatIDStr, err)
				return 0, ""
			}

			// Get message by removing @sendto and chatID
			message := text[len(parts[0])+len(parts[1])+2:] // +2 for space and @sendto
			return chatID, message
		}
	}
	return 0, ""
}

// initSpecialChatIDs initializes the special chat IDs from environment variable
func initSpecialChatIDs() {
	specialChatIDsStr := os.Getenv("SPECIAL_CHAT_IDS")
	if specialChatIDsStr == "" {
		LogInfo("SPECIAL_CHAT_IDS environment variable is not set, no special chats configured")
		return
	}

	chatIDStrings := strings.Split(specialChatIDsStr, ",")
	specialChatIDs = make([]int64, 0, len(chatIDStrings))

	for _, chatIDStr := range chatIDStrings {
		chatIDStr = strings.TrimSpace(chatIDStr)
		if chatIDStr == "" {
			continue
		}

		var chatID int64
		if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil {
			LogError("Invalid chat ID format in SPECIAL_CHAT_IDS: %s", chatIDStr)
			continue
		}

		specialChatIDs = append(specialChatIDs, chatID)
	}

	LogInfo("Initialized %d special chat IDs", len(specialChatIDs))
}
