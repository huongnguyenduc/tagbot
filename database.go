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
	CREATE TABLE IF NOT EXISTS group_settings (
		chat_id BIGINT PRIMARY KEY,
		timezone TEXT DEFAULT 'Asia/Ho_Chi_Minh',
		poll_time TEXT DEFAULT '10:10',
		remind_time TEXT DEFAULT '12:15'
	);
	CREATE TABLE IF NOT EXISTS votes (
		chat_id BIGINT,
		user_id BIGINT,
		answer TEXT,
		created_at TIMESTAMPTZ DEFAULT now(),
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

func saveVote(chatID, userID int64, answer string) {
	query := `
	INSERT INTO votes (chat_id, user_id, answer)
	VALUES ($1, $2, $3)
	ON CONFLICT (chat_id, user_id) DO UPDATE SET
		answer = EXCLUDED.answer,
		created_at = now();
	`
	if _, err := db.Exec(query, chatID, userID, answer); err != nil {
		log.Println("Failed to save vote:", err)
	}
}

func getGroupSettings(chatID int64) (timezone, pollTime, remindTime string, err error) {
	query := `SELECT timezone, poll_time, remind_time FROM group_settings WHERE chat_id = $1`
	err = db.QueryRow(query, chatID).Scan(&timezone, &pollTime, &remindTime)
	if err == sql.ErrNoRows {
		// Insert default settings if none found
		_, err = db.Exec(`INSERT INTO group_settings (chat_id) VALUES ($1) ON CONFLICT DO NOTHING`, chatID)
		if err != nil {
			return "", "", "", err
		}
		// Use defaults
		return "Asia/Ho_Chi_Minh", "10:10", "12:15", nil
	}
	return
}

func updateGroupSetting(chatID int64, field, value string) error {
	validFields := map[string]bool{"timezone": true, "poll_time": true, "remind_time": true}
	if !validFields[field] {
		return fmt.Errorf("invalid setting field: %s", field)
	}

	query := fmt.Sprintf("UPDATE group_settings SET %s = $1 WHERE chat_id = $2", field)
	res, err := db.Exec(query, value, chatID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// Insert default row then update again
		_, err = db.Exec(`INSERT INTO group_settings (chat_id) VALUES ($1) ON CONFLICT DO NOTHING`, chatID)
		if err != nil {
			return err
		}
		_, err = db.Exec(query, value, chatID)
	}
	return err
}

func listVotes(chatID int64) string {
	query := `SELECT answer, COUNT(*) FROM votes WHERE chat_id = $1 GROUP BY answer`
	rows, err := db.Query(query, chatID)
	if err != nil {
		log.Println("Failed to query votes:", err)
		return "Failed to get votes."
	}
	defer rows.Close()

	counts := map[string]int{"IN": 0, "OUT": 0, "MAYBE": 0}
	for rows.Next() {
		var ans string
		var cnt int
		if err := rows.Scan(&ans, &cnt); err == nil {
			counts[ans] = cnt
		}
	}
	return fmt.Sprintf("Current vote counts:\n🟢 IN: %d\n🔴 OUT: %d\n🤔 MAYBE: %d", counts["IN"], counts["OUT"], counts["MAYBE"])
}
