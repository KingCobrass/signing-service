package logger

import (
	"log"
	"os"
	"signing-service/config"
)

var errorlog *os.File

// Logger is the global logger
var Logger *log.Logger

func init() {
	errorlog, err := os.OpenFile(config.Env.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	Logger = log.New(errorlog, "apilog: ", log.Lshortfile|log.LstdFlags)
}
