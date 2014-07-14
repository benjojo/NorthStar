package main

import (
	"code.google.com/p/go.crypto/ssh"
	"net"
)

type PeerPacket struct {
	Service  string
	Message  string
	Host     string
	Salt     string // Make sure that this is uniq, else ur packet is gonna get dropped.
	LastSeen int64
}

func HandleNorthStarChan(Chan ssh.Channel, nConn net.Conn) {
	logger.Printf("New peer %s", nConn.RemoteAddr().String())
	// First we need to add the host into the peer list.
	WriteChan := make(chan []byte)
	NewPeer := Peer{}
	NewPeer.Alive = true
	NewPeer.ApparentIP = nConn.RemoteAddr().String()
	NewPeer.Conn = nConn
	NewPeer.MessageChan = WriteChan
	GlobalPeerList.Add(&NewPeer)

	go NSConnWriteDrain(WriteChan, Chan, &NewPeer)
	go NSConnReadDrain(GlobalResvChan, Chan)
}

func NSConnWriteDrain(inbound chan []byte, Chan ssh.Channel, Owner *Peer) {
	for outgoing := range inbound {
		_, err := Chan.Write(outgoing)
		if err != nil {
			debuglogger.Printf("Connection Write Drain shutdown.")
			Owner.Alive = false // Make sure the connection is not left hanging around
			return
		}
		debuglogger.Printf("Writing to channel %d bytes", len(outgoing))
	}
}

func NSConnReadDrain(inbound chan []byte, Chan ssh.Channel) {

	buffer := make([]byte, 25565)

	for {
		amt, err := Chan.Read(buffer)
		if err != nil {
			debuglogger.Printf("Connection Read Drain shutdown.")
			return
		}
		debuglogger.Printf("Read from channel %d bytes", amt)
		inbound <- buffer[:amt]
	}
}
