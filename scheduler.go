package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func scheduleGroupTasks(chatID int64) {
	// Cancel existing schedule if any
	schedulersMu.Lock()
	if cancel, ok := schedulers[chatID]; ok {
		cancel()
		delete(schedulers, chatID)
	}
	schedulersMu.Unlock()

	timezone, pollTime, remindTime, err := getGroupSettings(chatID)
	if err != nil {
		log.Println("Failed to get group settings, using defaults:", err)
		timezone = "Asia/Ho_Chi_Minh"
		pollTime = "10:10"
		remindTime = "12:15"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Invalid timezone '%s' for chat %d, using default UTC\n", timezone, chatID)
		loc = time.UTC
	}

	ctx, cancel := context.WithCancel(context.Background())

	schedulersMu.Lock()
	schedulers[chatID] = cancel
	schedulersMu.Unlock()

	go runScheduledJob(ctx, pollTime, loc, func() {
		// Check if manual poll was already triggered today before sending automatic poll
		if !wasManualPollTriggeredToday(chatID, loc) {
			sendPoll(chatID)
		} else {
			log.Printf("Skipping automatic poll for chat %d - manual poll already triggered today", chatID)
		}
	})

	go runScheduledJob(ctx, remindTime, loc, func() {
		sendReminder(chatID)
	})
}

func runScheduledJob(ctx context.Context, hhmm string, loc *time.Location, job func()) {
	parts := strings.Split(hhmm, ":")
	if len(parts) != 2 {
		log.Println("Invalid time format for scheduling:", hhmm)
		return
	}
	hour, errH := strconv.Atoi(parts[0])
	minute, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		log.Println("Invalid hour/minute in schedule:", hhmm)
		return
	}

	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		select {
		case <-time.After(time.Until(next)):
			job()
		case <-ctx.Done():
			return
		}
	}
}

// Helper function to get today's date key for manual poll tracking
func getTodayKey(chatID int64, loc *time.Location) string {
	today := time.Now().In(loc).Format("2006-01-02")
	return fmt.Sprintf("%d:%s", chatID, today)
}

// Check if manual poll was already triggered today
func wasManualPollTriggeredToday(chatID int64, loc *time.Location) bool {
	manualPollTriggersMu.Lock()
	defer manualPollTriggersMu.Unlock()
	
	key := getTodayKey(chatID, loc)
	return manualPollTriggers[key]
}

// Mark that manual poll was triggered today
func markManualPollTriggered(chatID int64, loc *time.Location) {
	manualPollTriggersMu.Lock()
	defer manualPollTriggersMu.Unlock()
	
	key := getTodayKey(chatID, loc)
	manualPollTriggers[key] = true
	
	// Clean up old entries (older than 7 days) to prevent memory leak
	cleanupOldManualPollTriggers()
}

// Clean up manual poll triggers older than 7 days
func cleanupOldManualPollTriggers() {
	cutoff := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	
	for key := range manualPollTriggers {
		// Extract date from key format "chatID:YYYY-MM-DD"
		parts := strings.Split(key, ":")
		if len(parts) == 2 && parts[1] < cutoff {
			delete(manualPollTriggers, key)
		}
	}
}
