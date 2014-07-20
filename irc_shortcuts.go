package main

import (
	"fmt"
	"os"
)

func GenerateIRCMessage(code string, nodename string, username string, data string) string {
	return fmt.Sprintf(":%s %s %s %s\r\n", nodename, code, username, data)
}

func GenerateIRCMessageBin(code string, nodename string, username string, data string) []byte {
	return []byte(GenerateIRCMessage(code, nodename, username, data))
}

func GetWelcomePackets(IRCUsername string) []byte {

	hostname, e := os.Hostname()
	if e != nil {
		hostname = "Unknown"
	}

	pack := ""
	pack += GenerateIRCMessage(RplWelcome, hostname, IRCUsername, fmt.Sprintf(":Welcome to NorthStar@%s", hostname))
	pack += GenerateIRCMessage(RplYourHost, hostname, IRCUsername, fmt.Sprintf(":Host is: %s", hostname))
	pack += GenerateIRCMessage(RplCreated, hostname, IRCUsername, ":This server was first made on 31/06/2014")
	pack += GenerateIRCMessage(RplMyInfo, hostname, IRCUsername, fmt.Sprintf(":%s NS DOQRSZaghilopswz CFILMPQSbcefgijklmnopqrstvz bkloveqjfI", hostname))
	pack += GenerateIRCMessage(RplMotdStart, hostname, IRCUsername, ":There is no MOTD.")
	pack += GenerateIRCMessage(RplMotdEnd, hostname, IRCUsername, ":done")
	return []byte(pack)
}

func GenerateIRCPrivateMessage(content string, room string, username string, nodename string) []byte {
	return []byte(fmt.Sprintf(":%s!~%s@%s PRIVMSG %s :%s\r\n", username, username, nodename, room, content))
}
