package main

import (
	"log"
	"os"
)

var logger *log.Logger

func SetupLogger() {
	logger = log.New(os.Stdout, "[Main] ", log.Lshortfile|log.Ltime)
}
