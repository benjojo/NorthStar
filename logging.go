package main

import (
	"io/ioutil"
	"log"
	"os"
)

var logger *log.Logger
var debuglogger *log.Logger

func SetupLogger(debug bool) {
	logger = log.New(os.Stdout, "[Main] ", log.Lshortfile|log.Ltime)
	if debug {
		logger = log.New(os.Stdout, "[Debug] ", log.Lshortfile|log.Ltime)
	} else {
		logger = log.New(ioutil.Discard, "[Debug] ", log.Lshortfile|log.Ltime)
	}
}
