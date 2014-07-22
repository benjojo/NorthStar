package main

import (
	"flag"
	"io/ioutil"
	"time"
)

var CC_KEY string = ""
var PEM_KEY []byte

func main() {
	key := flag.String("key", GetFileOrBlank("/.nskey"), "Set this as a long string that only you and the rest of your nodes know")
	seed := flag.String("seed", "", "Set this to the key of the network, Need atleast 1 matchine on the network to have this")
	dhtpos := flag.Bool("dhtdefault", false, "If you cannot get announcing to work, then set this to true")
	debug := flag.Bool("debug", false, "Enable for debug text")
	flag.Parse()

	SetupLogger(*debug)
	go ListenForIRCConnections()

	if *key == "" || *key == "InsertLongKeyHere" || len(*key) < 8 {
		logger.Fatal("No key or a too short key use -key=\"InsertLongKeyHere\"")
	}
	CC_KEY = string(HashValue([]byte(*key))[:16])
	ReadyToBroadcast = *dhtpos
	logger.Printf("NorthStar - DHT Configuration Tool.")
	go StartLookingForPeers()

	if *seed != "" {
		// Okay we are going into preseeded mode!
		PEM_KEY = LoadPrivKeyFromFile(*seed)
	} else {
		PEM_KEY = PullDHTKey()
	}

	private, err := ssh.ParsePrivateKey(PEM_KEY)
	if err != nil {
		logger.Fatal("Key failed to parse.")
	}
	GSigner = private

	ReadyToBroadcast = true
	logger.Printf("Got the key.")
	go WaitForConnections()
	go ScountOutNewPeers()
	go RelayPackets()
	for {
		time.Sleep(time.Second * 10)
		Holla := PeerPacket{}
		Holla.Message = "Hi everyone"
		Holla.Service = "Holla"
		SendPacket(Holla)
	}
}

func GetFileOrBlank(path string) string {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		return ""
	}
	return string(b)
}
