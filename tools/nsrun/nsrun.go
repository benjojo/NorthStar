package main

import (
	"bufio"
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	keypath := flag.String("key", "./.nsmasterkey", "The path to the NorthStar master key")
	extra := flag.Args()
	flag.Parse()
	cmdlinestring := ""
	con, err := net.Dial("tcp", "localhost:6669")
	connreader := bufio.NewReader(con)

	if err != nil {
		log.Fatal("Unable to connect to northstar, is it online?")
	}

	if len(extra) != 0 {
		// The command is in there!
		for _, v := range extra {
			cmdlinestring = cmdlinestring + v + " "
		}
	}

	if cmdlinestring == "" {
		fmt.Print("> ")
		rd := bufio.NewReader(os.Stdin)
		var err error
		cmdlinestringb, _, err := rd.ReadLine()
		cmdlinestring = string(cmdlinestringb)
		if err != nil {
			log.Fatalf("Unable to read promt %s", err.Error())
		}
	}

	con.Write([]byte("NICK NSRUN NSRUN NSRUN\r\n"))
	Unlock := MakeKey(*keypath)
	for {
		line, _, err := connreader.ReadLine()
		if err != nil {
			log.Fatalf("Unable to read handshake %s", err.Error())
		}

		if strings.Contains(string(line), " 376 ") {
			break
		}
	}
	// Handshake done.

	con.Write([]byte("JOIN #RPC\r\n"))
	for {
		line, _, err := connreader.ReadLine()
		if err != nil {
			log.Fatalf("Unable to read join %s", err.Error())
		}

		if strings.Contains(string(line), "JOIN #RPC *") {
			break
		}
	}
	con.Write([]byte(fmt.Sprintf("PRIVMSG #RPC :!%s\r\n", Unlock)))
	ExpectedResponces := make([]string, 0)
	NewLinesChan := make(chan string)
	go ReadLinesOffConnection(connreader, NewLinesChan)
CommsLoop:
	for {
		select {
		case line := <-NewLinesChan:
			if strings.Contains(line, "System Unlocked") {
				ExpectedResponces = AddHostsToExpectedList(line, ExpectedResponces)
			}
		case <-time.After(time.Second * 5):
			break CommsLoop
		}
	}

	// Print out the hosts who we are expecting responces from.
	fmt.Print("Systems who responded: ")
	for _, v := range ExpectedResponces {
		fmt.Print(v + " ")
	}
	fmt.Print("\n")

	con.Write([]byte(fmt.Sprintf("PRIVMSG #RPC :#%s\r\n", cmdlinestring)))

	Responces := make(map[string][]string)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

WaitForResponcesLoop:
	for {
		select {
		case line := <-NewLinesChan:
			if strings.Contains(line, "PRIVMSG #RPC :") {
				// A responce!
				HandlePossibleResponce(line, Responces)
			}
		case <-time.After(time.Second * 60):
			break WaitForResponcesLoop
		case <-c:
			break WaitForResponcesLoop
		}
	}

	for h, responce := range Responces {
		fmt.Println("-----------------------------------------")
		fmt.Printf("Host: %s\n\n", h)
		for _, v := range responce {
			fmt.Println(v)
		}
	}
}

func HandlePossibleResponce(inbound string, datamap map[string][]string) {
	// :au!~au@au PRIVMSG #RPC :~5mF2x [1/1] -  23:08:13 up 57 days, 20:58,  0 users,  load average: 0.00, 0.00, 0.00
	hostname := inbound[1:strings.Index(inbound, "!")]
	bits := strings.Split(inbound, " ")
	// take apart the [1/1] part
	progress := strings.Split(bits[4][1:strings.Index(bits[4], "]")], "/")

	cur, err := strconv.ParseInt(progress[0], 10, 64)
	if err != nil {
		log.Fatal("Unable to decode number !!!")
	}
	max, err := strconv.ParseInt(progress[1], 10, 64)
	if err != nil {
		log.Fatal("Unable to decode number !!!")
	}

	if datamap[hostname] == nil {
		datamap[hostname] = make([]string, max+1)
	}

	CleanString := strings.Replace(inbound, fmt.Sprintf("%s %s %s %s %s %s", bits[0], bits[1], bits[2], bits[3], bits[4], bits[5]), "", 1)
	datamap[hostname][cur-1] = CleanString
}

func ReadLinesOffConnection(in *bufio.Reader, out chan string) {
	for {
		line, _, err := in.ReadLine()
		if err != nil {
			log.Fatal(err.Error())
		}
		if !strings.Contains(string(line), "PING :1") {
			out <- string(line)
		}
	}
}

func AddHostsToExpectedList(IRCMessage string, list []string) []string {
	if strings.HasPrefix(IRCMessage, ":") && strings.Contains(IRCMessage, "PRIVMSG #RPC :System Unlocked.") {
		// Okay so a host has unlocked!
		hostname := IRCMessage[1:strings.Index(IRCMessage, "!")]
		list = append(list, hostname)
		return list
	}
	return list
}

func MakeKey(keypath string) string {
	key := LoadPrivKeyFromFile(keypath)

	private, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Key failed to parse. - %s", err.Error())
	}

	unixtimenow := time.Now().Unix()

	signtext := fmt.Sprintf("%d %d", unixtimenow, 120)
	sig, err := private.Sign(rand.Reader, []byte(signtext))
	if err != nil {
		log.Fatalf("Key failed to sign. - %s", err.Error())
	}

	hex := hex.EncodeToString(sig.Blob)
	return fmt.Sprintf("%d %d %s", unixtimenow, 120, hex)
}

func LoadPrivKeyFromFile(file string) []byte {
	privateBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Failed to load private key - %s", err.Error())
	}
	return privateBytes
}
