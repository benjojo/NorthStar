package main

import (
	"fmt"
	"github.com/benjojo/dht"
	"strings"
	"time"
)

// StartDHT will start the DHT (only to be called once) and then give back a chan
// that all of the peer responces will be given back on in the format "1.1.1.1:222"
// beware that the command is filtered for peers that have the port 48563
func StartDHT(infohash string) chan string {
	ih, err := dht.DecodeInfoHash(infohash)

	if err != nil {
		logger.Fatalf("DHT cannot decode a infohash to find peers on, %s", err)
	}
	config := dht.NewConfig()
	config.Port = 48563 // int(md5("northstar")[:2])
	d, err := dht.New(config)
	if err != nil {
		logger.Fatalf("DHT cannot start, %s", err)
	}
	logger.Printf("Starting DHT on port %d now...", config.Port)
	go d.Run()
	go DHTBroadcast(d, ih)
	outputchan := make(chan string)
	go DHTDrain(outputchan, d)
	return outputchan
}

var ReadyToBroadcast bool = false

func DHTBroadcast(d *dht.DHT, hash dht.InfoHash) {
	for {
		d.PeersRequest(string(hash), ReadyToBroadcast)
		time.Sleep(time.Second * 30)
	}
}

func DHTDrain(out chan string, d *dht.DHT) {
	for r := range d.PeersRequestResults {
		for _, peers := range r {
			for _, x := range peers {
				decodedaddr := dht.DecodePeerAddress(x)
				if strings.Split(decodedaddr, ":")[1] == fmt.Sprint(d.Port()) {
					// Ding!
					out <- decodedaddr
					logger.Printf("Found a peer though DHT (%s)", decodedaddr)
				}
			}
		}
	}
}
