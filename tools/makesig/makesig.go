package main

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	keypath := flag.String("key", "", "Where the private key is")
	windowtime := flag.Int("window", 120, "How large the window is to run commands")
	flag.Parse()
	key := LoadPrivKeyFromFile(*keypath)

	private, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Key failed to parse. - %s", err.Error())
	}

	unixtimenow := time.Now().Unix()

	signtext := fmt.Sprintf("%d %d", unixtimenow, *windowtime)
	sig, err := private.Sign(rand.Reader, []byte(signtext))
	if err != nil {
		log.Fatalf("Key failed to sign. - %s", err.Error())
	}

	hex := hex.EncodeToString(sig.Blob)
	fmt.Printf("%d %d %s", unixtimenow, *windowtime, hex)
}

func LoadPrivKeyFromFile(file string) []byte {
	privateBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Failed to load private key - %s", err.Error())
	}
	return privateBytes
}
