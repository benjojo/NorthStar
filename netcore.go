package main

import (
	"net"
)

func WatchForNewPeers(inbound chan string) {

}

func WaitForConnections() {
	l, err := net.Listen("tcp", "0.0.0.0:48563")
	if err != nil {
		logger.Fatalf("Could not start listening on TCP side (%s)", err)
	}
	defer l.Close()

	logger.Println("TCP Side is ready for connections.")

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			logger.Println("Error accepting TCP: ", err.Error())
		}
		// Handle connections in a new goroutine.
		go HandlePeerConn(conn)
	}
}

func HandlePeerConn(conn net.Conn) {
	err := ChallengeClient(conn)
	if err != nil {
		conn.Close()
	}
	defer conn.Close()
	payload := make([]byte, 64000)
	for {
		conn.Read(payload)
	}
}
