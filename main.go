package main

import (
	"flag"
)

func main() {
	SetupLogger()
	key := flag.String("key", "", "Set this as a long string that only you and the rest of your nodes know")
	flag.Parse()
	if *key == "" || *key == "InsertLongKeyHere" {
		logger.Fatal("No key or a too short key use -key=\"InsertLongKeyHere\"")
	}

	logger.Printf("NorthStar - DHT Configuration Tool.")
}
