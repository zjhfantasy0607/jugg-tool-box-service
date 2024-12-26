package util

import (
	"log"
	"os"
)

func Logln(path string, v ...interface{}) {
	// Open the log file in append mode or create it if it doesn't exist
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Create a new logger that writes to the file
	logger := log.New(file, "", log.LstdFlags)

	// Log the message
	logger.Println(v...)
}

func LogErr(myerr error, path string) {
	// Open the log file in append mode or create it if it doesn't exist
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Create a new logger that writes to the file
	logger := log.New(file, "", log.LstdFlags)
	logger.Printf("\nError: %+v\n\n", myerr)
}
