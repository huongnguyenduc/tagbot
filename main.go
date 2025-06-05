package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

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

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./members.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS members (
		chat_id INTEGER,
		user_id INTEGER,
		first_name TEXT,
		last_name TEXT,
		username TEXT,
		PRIMARY KEY (chat_id, user_id)
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func saveUser(chatID int64, user *tgbotapi.User) {
	stmt, err := db.Prepare(`
	INSERT OR REPLACE INTO members (chat_id, user_id, first_name, last_name, username)
	VALUES (?, ?, ?, ?, ?);
	`)
	if err != nil {
		log.Println("Prepare failed:", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(chatID, user.ID, user.FirstName, user.LastName, user.UserName)
	if err != nil {
		log.Println("Insert failed:", err)
	}
}

func deleteUser(chatID int64, userID int64) {
	_, err := db.Exec("DELETE FROM members WHERE chat_id = ? AND user_id = ?", chatID, userID)
	if err != nil {
		log.Println("Delete failed:", err)
	}
}

func getMentions(chatID int64) string {
	rows, err := db.Query(`SELECT user_id, first_name, last_name, username FROM members WHERE chat_id = ?`, chatID)
	if err != nil {
		log.Println("Query failed:", err)
		return ""
	}
	defer rows.Close()

	var mentions []string
	for rows.Next() {
		var userID int
		var firstName, lastName, username string
		err = rows.Scan(&userID, &firstName, &lastName, &username)
		if err != nil {
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

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Missing TELEGRAM_BOT_TOKEN environment variable")
	}

	initDB()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	// Delete webhook if it exists
	if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		log.Fatalf("Failed to delete webhook: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message
		chatID := msg.Chat.ID

		// Save user when they send a message
		if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
			if msg.From != nil {
				saveUser(chatID, msg.From)
			}
		}

		// Handle user leaving the group
		if msg.LeftChatMember != nil {
			deleteUser(chatID, msg.LeftChatMember.ID)
			log.Printf("User %d left chat %d and was removed from DB", msg.LeftChatMember.ID, chatID)
			continue
		}

		// Handle @all mentions
		if (msg.Chat.IsGroup() || msg.Chat.IsSuperGroup()) && strings.Contains(strings.ToLower(msg.Text), "@all") {
			messageText := escapeMarkdownV2(removeAtAll(msg.Text))
			tags := getMentions(chatID)

			if tags == "" {
				log.Println("No members found to tag.")
				continue
			}

			fullText := fmt.Sprintf("%s %s", messageText, tags)

			reply := tgbotapi.NewMessage(chatID, fullText)
			reply.ParseMode = "MarkdownV2"
			reply.ReplyToMessageID = msg.MessageID

			if _, err := bot.Send(reply); err != nil {
				log.Println("Send error:", err)
			}
		}
	}
}
