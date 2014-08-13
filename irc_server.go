package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var ConnectedIRCClients IRCClients

type IRCClients struct {
	Clients map[int]*IRCClient
	m       sync.Mutex
}

type IRCClient struct {
	Channels    map[string]bool
	InboundChan chan PeerPacket
}

func ListenForIRCConnections() {
	ConnectedIRCClients.Clients = make(map[int]*IRCClient)
	// Listen for incoming connections.
	l, err := net.Listen("tcp", "localhost:6669")
	if err != nil {
		logger.Fatalf("Unable to bind IRC server.")
	}
	// Close the listener when the application closes.
	defer l.Close()
	logger.Println("IRCServer Listening on localhost:6669")
	var IRCConNumber int
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			logger.Println("Error accepting: ", err.Error())
		}
		// Handle connections in a new goroutine.
		go HandleIRCConn(conn, IRCConNumber)
		IRCConNumber++
	}
}

func HandleIncomingIRCChanMsgs(in chan PeerPacket, conn net.Conn) {
	for msg := range in {
		_, err := conn.Write(GenerateIRCPrivateMessage(msg.Message, "#"+msg.Service, msg.Host, msg.Host))
		if err != nil {
			close(in)
			return
		}
	}
}

// Handles incoming requests.
func HandleIRCConn(conn net.Conn, connectionnumber int) {
	ConnectedIRCClients.m.Lock()

	MySelf := IRCClient{}
	MySelf.InboundChan = make(chan PeerPacket)
	MySelf.Channels = make(map[string]bool)
	ConnectedIRCClients.Clients[connectionnumber] = &MySelf
	ConnectedIRCClients.m.Unlock()

	defer delete(ConnectedIRCClients.Clients, connectionnumber)
	go HandleIncomingIRCChanMsgs(MySelf.InboundChan, conn)
	reader := bufio.NewReader(conn)
	IRCUsername := "BrokenIRCClient"
	go PingIRCClient(conn)

	for {
		lineb, _, err := reader.ReadLine()
		if err != nil {
			conn.Close()
			return
		}
		line := string(lineb)

		if strings.HasPrefix(line, "QUIT ") {
			conn.Close()
			return
		}

		if strings.HasPrefix(line, "NICK ") {
			IRCUsername = strings.Split(line, " ")[1]
			conn.Write(GetWelcomePackets(IRCUsername))
		}

		if strings.HasPrefix(line, "JOIN ") {
			ChanName := strings.Replace(strings.Split(line, " ")[1], "#", "", -1)
			MySelf.Channels[ChanName] = true
			conn.Write([]byte(fmt.Sprintf(":%s!~%s@localhost JOIN #%s * :NS Client\r\n", IRCUsername, IRCUsername, ChanName)))
		}

		if strings.HasPrefix(line, "PART ") {
			ChanName := strings.Replace(strings.Split(line, " ")[1], "#", "", -1)
			MySelf.Channels[ChanName] = false
			conn.Write([]byte(fmt.Sprintf(":%s!~%s@localhost PART #%s * :NS Client\r\n", IRCUsername, IRCUsername, ChanName)))
		}

		if strings.HasPrefix(line, "PRIVMSG ") {
			ChanName := strings.Replace(strings.Split(line, " ")[1], "#", "", -1)
			//HACKS ALERT
			Message := strings.Replace(line, "PRIVMSG #"+ChanName+" :", "", -1)
			Outbound := PeerPacket{}
			Outbound.Service = ChanName
			Outbound.Message = Message
			SendPacket(Outbound)
		}
	}
}

func PingIRCClient(conn net.Conn) {
	for {
		_, e := conn.Write([]byte(fmt.Sprintf("PING :%d\r\n", int32(time.Now().Unix()))))
		if e != nil {
			break
		}
		time.Sleep(time.Second * 30)
	}
}