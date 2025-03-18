package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
)

var Logger *log.Logger

func Init123() error {
	// Get the current timestamp in the desired format (with milliseconds)
	timestamp := time.Now().Format("2006-01-02_15-04-05.000")

	// Create a folder with the timestamp
	logDir := "exe-logs/run_" + timestamp
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return err
	}

	// Configure log rotation with lumberjack
	logFile := &lumberjack.Logger{
		Filename:   logDir + "/orderpumping.log", // Log file name with timestamped folder
		MaxSize:    1024,                         // Max size before rotation (in MB)
		MaxBackups: 50,                           // Max number of backups to keep
		MaxAge:     30,                           // Max number of days to retain backups
		Compress:   true,                         // Compress old log files
	}

	log.SetOutput(logFile)

	Logger = log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
	return nil
}

func Printf(flag bool, statement string) {
	if flag {
		fmt.Printf(statement)
	}
}
