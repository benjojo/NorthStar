package main

import (
	"errors"
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

func ChallengeClient(conn net.Conn) (err error) {
	Challenge := RandString(32)
	conn.Write(EncryptText([]byte("NS "+Challenge), []byte(CC_KEY)))
	Responce := make([]byte, 64)
	back, e := conn.Read(Responce)
	if e != nil {
		return e
	}
	decrypted := DecryptText(Responce[:back], []byte(CC_KEY))
	if string(decrypted) == Challenge {
		return err
	} else {
		return errors.New("AAAAAAAAAA")
	}
}

/*

Connect
S->C "NS"
C->S "{Cypto Blob containing Time and RandomString}"
S->C "{Cypto Blob containing the hashed version of the random string and the time}"
C->S "{ACK packet with hashed version of the decrypted GOB}"
S->C "{Ask for Hostname and other info}"
C->S "{GOB containing host info}"
-- Normal Relay mode starts --
*/
