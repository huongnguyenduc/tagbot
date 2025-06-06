package main

import (
	"context"
	"database/sql"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	db  *sql.DB
	bot *tgbotapi.BotAPI

	// Map to hold cancel funcs for scheduled jobs per chatID
	schedulers   = make(map[int64]context.CancelFunc)
	schedulersMu sync.Mutex

	// Map to track manual poll triggers per chat per day
	manualPollTriggers   = make(map[string]bool) // key: "chatID:YYYY-MM-DD"
	manualPollTriggersMu sync.Mutex
)
