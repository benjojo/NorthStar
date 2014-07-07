package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math"
	"net"
	"time"
)

// To be used when client is connecting TO a server
func ChallengeClient(conn net.Conn) (err error) {
	Challenge := RandString(32)
	conn.Write(EncryptText([]byte("NS"), []byte(CC_KEY)))
	Responce := make([]byte, 1024)
	back, e := conn.Read(Responce)
	if e != nil {
		return e
	}
	decrypted := DecryptText(Responce[:back], []byte(CC_KEY))
	var clearnet bytes.Buffer
	clearnet.Write(decrypted) // Write the decrypted blob into the nic
	gobdec := gob.NewDecoder(&clearnet)
	gobenc := gob.NewEncoder(&clearnet)

	HSB := HandShakePack{}
	err = gobdec.Decode(&HSB)
	if err != nil {
		return errors.New("Handshake failed, could not decode HS")
	}

	// Test the time lag to ensure its not a listen and repeat attack
	if math.Abs(float64(time.Now().Unix()-HSB.ChallengeTime)) < 1000 {
		return errors.New("Handshake failed, time lag between packet was too high")
	}

	// This is proabbly a good idea to check anyway
	if HSB.ChallengeString == "" {
		return errors.New("Handshake failed, there was no challenge string")
	}
	NewChallenge := HSB.ChallengeString

	// Okay so this packet passes the valid test
	// Now to make the responce to that packet.
	HSB = HandShakePack{}
	HSB.ChallengeTime = time.Now().Unix()
	HSB.ChallengeResponce = string(HashValue([]byte(NewChallenge)))
	HSB.ChallengeString = Challenge // Just so the client does not fiddle with us.
	gobenc.Encode(&HSB)
	tocrypt := clearnet.Bytes()
	upcomingchallenge := HashValue(tocrypt)
	tosend := EncryptText(tocrypt, []byte(CC_KEY))
	_, err = conn.Write(tosend)
	if err != nil {
		return errors.New("Handshake failed, Cannot write 2nd stage responce to NIC")
	}
	back, err = conn.Read(Responce)
	if err != nil {
		return errors.New("Handshake failed, Cannot read 3nd stage responce from client")
	}

	decrypted = DecryptText(Responce[:back], []byte(CC_KEY))
	HSB = HandShakePack{}
	err = gobdec.Decode(&HSB)
	if err != nil {
		return errors.New("Handshake failed, could not decode 3rd stage HS")
	}
	// Test the time lag to ensure its not a listen and repeat attack
	if math.Abs(float64(time.Now().Unix()-HSB.ChallengeTime)) < 1000 {
		return errors.New("Handshake failed, time lag between packet was too high")
	}

	if HSB.ChallengeResponce != string(upcomingchallenge) {
		return errors.New("Handshake failed, Final hash was wrong!!!")
	}
	return err // Well it looks like this guy is legit.
}

// To be used when connecting TO a server
func ChallengeServer(conn net.Conn) (err error) {
	//Responce := make([]byte, 1024)
	//back, e := conn.Read(Responce)
	//if e != nil {
	//	return e
	//}
	//decrypted := DecryptText(Responce[:back], []byte(CC_KEY))
	//if string(decrypted) != "NS" {
	//	return errors.New("Handshake failed, Banner was incorrect")
	//}
	//// Check one done.
	//// Now to setup the GOB encoders
	//var clearnet bytes.Buffer
	//gobdec := gob.NewDecoder(&clearnet)
	//gobenc := gob.NewEncoder(&clearnet)
	//
	//HSB := HandShakePack{}
	//Challenge := RandString(32)
	//Expectedresponce := HashValue([]byte(Challenge))
	//HSB.ChallengeString = Challenge
	//HSB.ChallengeTime = time.Now().Unix()
	//
	//// Encode and crypt
	//gobenc.Encode(&HSB)
	//tocrypt := clearnet.Bytes()
	//tosend := EncryptText(tocrypt, []byte(CC_KEY))
	//conn.Write(tosend)

	return err // Well it looks like this guy is legit.
}

type HandShakePack struct {
	ChallengeString   string
	ChallengeResponce string
	ChallengeTime     int64
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

Wait a sec, why not just use SSH?

*/
