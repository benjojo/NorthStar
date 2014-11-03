package main

import (
	"encoding/json"
	"strings"
	"time"
)

type PEXPacket struct {
	Peers []string
}

var LastPEXTime int64 = 0
var PeerAtlas map[string][]string

func ProcessPEXPacket(P PeerPacket) {
	if PeerAtlas == nil {
		PeerAtlas = make(map[string][]string)
	}

	PEXData := PEXPacket{}
	err := json.Unmarshal([]byte(P.Message), &PEXData)
	if err != nil {
		logger.Printf("Invalid PEX packet sent from %s", P.Host)
		return
	}

	for i := 0; i < len(PEXData.Peers); i++ {
		if !GlobalPeerList.ContainsIP(PEXData.Peers[i]) {
			NewPeer := Peer{}
			NewPeer.Alive = false
			NewPeer.ApparentIP = PEXData.Peers[i]
			GlobalPeerList.Add(&NewPeer, -1)
			debuglogger.Printf("Obtained peer (%s) though PEX thanks to %s", PEXData.Peers[i], P.Host)
		}
	}
	PeerAtlas[P.Host] = PEXData.Peers
}

func MakePEXPacket() {

	if LastPEXTime+30 > time.Now().Unix() { // Only allow PEX every 30 secs
		return
	}
	LastPEXTime = time.Now().Unix()

	Outgoing := PEXPacket{}
	Outgoing.Peers = make([]string, 0)

	GlobalPeerList.m.Lock()
	debuglogger.Println("GPList is locked")
	Seen := make(map[string]bool)
	for _, v := range GlobalPeerList.Peers {
		if v.Alive && !Seen[v.ApparentIP] {
			PeerIP := v.ApparentIP
			PeerIP = strings.Split(PeerIP, ":")[0]
			PeerIP = PeerIP + ":48563"
			Outgoing.Peers = append(Outgoing.Peers, PeerIP)
			Seen[v.ApparentIP] = true
		}
	}
	debuglogger.Println("GPList is unlocked")
	GlobalPeerList.m.Unlock()

	Packet := PeerPacket{}
	Packet.Service = "PEX"
	b, err := json.Marshal(&Outgoing)
	if err != nil {
		logger.Printf("Could not encode PEX packet (WTF)")
	}
	Packet.Message = string(b)
	SendPacket(Packet)
}

func AskForPEX() {
	P := PeerPacket{}
	P.Service = "PEX_REQUEST"
	P.Message = "PEX Please"
	SendPacket(P)
}
