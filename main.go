package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
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
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Missing DATABASE_URL environment variable")
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS members (
		chat_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		first_name TEXT,
		last_name TEXT,
		username TEXT,
		PRIMARY KEY (chat_id, user_id)
	);
	`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func saveUser(chatID int64, user *tgbotapi.User) {
	stmt, err := db.Prepare(`
	INSERT INTO members (chat_id, user_id, first_name, last_name, username)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (chat_id, user_id) DO UPDATE
	SET first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		username = EXCLUDED.username;
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
	_, err := db.Exec("DELETE FROM members WHERE chat_id = $1 AND user_id = $2", chatID, userID)
	if err != nil {
		log.Println("Delete failed:", err)
	}
}

func getMentions(chatID int64) string {
	rows, err := db.Query(`SELECT user_id, first_name, last_name, username FROM members WHERE chat_id = $1`, chatID)
	if err != nil {
		log.Println("Query failed:", err)
		return ""
	}
	defer rows.Close()

	var mentions []string
	for rows.Next() {
		var userID int64
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

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Setup webhook
	webhookURL := os.Getenv("WEBHOOK_URL") // e.g. https://yourdomain.com/telegram-webhook
	if webhookURL == "" {
		log.Fatal("Missing WEBHOOK_URL environment variable")
	}

	whcfg, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		log.Fatal("NewWebhook failed:", err)
	}

	_, err = bot.Request(whcfg)
	if err != nil {
		log.Fatal("Failed to set webhook:", err)
	}

	// Get webhook info
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal("Failed to get webhook info:", err)
	}
	if info.URL != webhookURL {
		log.Fatalf("Webhook URL mismatch. Got: %s", info.URL)
	}

	log.Printf("Webhook set to %s", webhookURL)

	// Start HTTP server to receive updates from Telegram
	http.HandleFunc("/telegram-webhook", func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			log.Println("Failed to handle update:", err)
			return
		}

		if update.Message == nil {
			return
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
			return
		}

		// Handle @all mentions
		if (msg.Chat.IsGroup() || msg.Chat.IsSuperGroup()) && strings.Contains(strings.ToLower(msg.Text), "@all") {
			messageText := escapeMarkdownV2(removeAtAll(msg.Text))
			tags := getMentions(chatID)

			if tags == "" {
				log.Println("No members found to tag.")
				return
			}

			fullText := fmt.Sprintf("%s %s", messageText, tags)

			reply := tgbotapi.NewMessage(chatID, fullText)
			reply.ParseMode = "MarkdownV2"
			reply.ReplyToMessageID = msg.MessageID

			if _, err := bot.Send(reply); err != nil {
				log.Println("Send error:", err)
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
