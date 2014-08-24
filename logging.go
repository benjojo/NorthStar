package main

import (
	"io/ioutil"
	"log"
	"os"
)

var logger *log.Logger
var debuglogger *log.Logger
var droppedlogger *log.Logger

func SetupLogger(debug bool) {
	logger = log.New(os.Stdout, "[Main] ", log.Lshortfile|log.Ltime)
	if debug {
		debuglogger = log.New(os.Stdout, "[Debug] ", log.Lshortfile|log.Ltime)
	} else {
		debuglogger = log.New(ioutil.Discard, "[Debug] ", log.Lshortfile|log.Ltime)
	}

	if LogDroppedPackets {
		dl, err := os.Open("./droppedpackets")
		if err != nil {
			logger.Fatalf("Could not open a file handle for the droplog RSN: %s", err.Error())
		}
		droppedlogger = log.New(dl, "[Dropped] ", log.Ltime)
	}
}
