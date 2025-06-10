package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func initLogger() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Create or open log file with date in filename
	currentTime := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(
		fmt.Sprintf("logs/tagbot-%s.log", currentTime),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Initialize loggers
	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo logs informational messages
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	infoLogger.Output(2, msg)
	log.Printf("INFO: %s", msg)
}

// LogError logs error messages
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	errorLogger.Output(2, msg)
	log.Printf("ERROR: %s", msg)
}

// LogFatal logs fatal errors and exits the program
func LogFatal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	errorLogger.Output(2, msg)
	log.Fatalf("FATAL: %s", msg)
}
