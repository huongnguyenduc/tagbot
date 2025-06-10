package main

import (
	"database/sql"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		LogFatal("Missing DATABASE_URL environment variable")
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		LogFatal("Failed to open DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		LogFatal("Failed to ping DB: %v", err)
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
		LogFatal("Failed to create tables: %v", err)
	}
	LogInfo("Database initialized successfully")
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
		LogError("Failed to save user %d in chat %d: %v", user.ID, chatID, err)
	} else {
		LogInfo("Saved user %d in chat %d", user.ID, chatID)
	}
}

func deleteUser(chatID int64, userID int64) error {
	if _, err := db.Exec("DELETE FROM members WHERE chat_id = $1 AND user_id = $2", chatID, userID); err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	LogInfo("Deleted user %d from chat %d", userID, chatID)
	return nil
}
