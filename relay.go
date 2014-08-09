package main

import (
	"code.google.com/p/go.crypto/ssh"
	"net"
	"time"
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
	go NSConnReadDrain(GlobalResvChan, Chan, &NewPeer)
}

func NSConnWriteDrain(inbound chan []byte, Chan ssh.Channel, Owner *Peer) {
	for outgoing := range inbound {
		_, err := Chan.Write(outgoing)
		if err != nil {
			debuglogger.Printf("Connection Write Drain shutdown.")
			Owner.Alive = false // Make sure the connection is not left hanging around
			close(Owner.MessageChan)
			return
		}
		debuglogger.Printf("Writing to channel %d bytes", len(outgoing))
	}
}

func NSConnReadDrain(inbound chan []byte, Chan ssh.Channel, Owner *Peer) {

	buffer := make([]byte, 25565)
	var ReadLimitTime int64
	var PacketsRead int

	for {
		amt, err := Chan.Read(buffer)
		if err != nil {
			debuglogger.Printf("Connection Read Drain shutdown.")
			Owner.Alive = false
			close(Owner.MessageChan)
			return
		}
		debuglogger.Printf("Read from channel %d bytes", amt)
		inbound <- buffer[:amt]
		PacketsRead++
		if PacketsRead > PacketRateLimit {
			logger.Printf("Rate limit kicked in for %s This is a sign of heavy traffic of bugs", Owner.ApparentIP)
			time.Sleep(time.Millisecond * 100)
		}
		if ReadLimitTime != time.Now().Unix() {
			PacketsRead = 0
			ReadLimitTime = time.Now().Unix()
		}
	}
}
