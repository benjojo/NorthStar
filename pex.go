package main

import (
	"encoding/json"
	"strings"
)

type PEXPacket struct {
	Peers []string
}

func ProcessPEXPacket(P PeerPacket) {
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
			GlobalPeerList.Add(&NewPeer)
			debuglogger.Printf("Obtained peer (%s) though PEX thanks to %s", PEXData.Peers[i], P.Host)
		}
	}
}

func MakePEXPacket() {
	Outgoing := PEXPacket{}
	Outgoing.Peers = make([]string, 0)

	GlobalPeerList.m.Lock()
	for _, v := range GlobalPeerList.Peers {
		if v.Alive {
			PeerIP := v.ApparentIP
			PeerIP = strings.Split(PeerIP, ":")[0]
			PeerIP = PeerIP + ":48563"
			Outgoing.Peers = append(Outgoing.Peers, PeerIP)
		}
	}
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