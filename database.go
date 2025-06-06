package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

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

	// Create tables if not exist
	schema := `
	CREATE TABLE IF NOT EXISTS members (
		chat_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		first_name TEXT,
		last_name TEXT,
		username TEXT,
		PRIMARY KEY (chat_id, user_id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal("Failed to create tables:", err)
	}
}

func saveUser(chatID int64, user *tgbotapi.User) {
	query := `
	INSERT INTO members (chat_id, user_id, first_name, last_name, username)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (chat_id, user_id) DO UPDATE SET
		first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		username = EXCLUDED.username;
	`
	if _, err := db.Exec(query, chatID, user.ID, user.FirstName, user.LastName, user.UserName); err != nil {
		log.Println("Insert user failed:", err)
	}
}

func deleteUser(chatID int64, userID int64) error {
	if _, err := db.Exec("DELETE FROM members WHERE chat_id = $1 AND user_id = $2", chatID, userID); err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}
