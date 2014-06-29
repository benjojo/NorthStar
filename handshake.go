package main

import (
	"errors"
	"net"
)

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
