package main

import (
	"database/sql"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	db  *sql.DB
	bot *tgbotapi.BotAPI
)
