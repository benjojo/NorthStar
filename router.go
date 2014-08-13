package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

var GlobalResvChan chan []byte
var PacketCache map[int]*PeerPacket

func RelayPackets() {
	GlobalResvChan = make(chan []byte)
	PacketCache = make(map[int]*PeerPacket)
	RouterChan := make(chan PeerPacket)
	go PacketRouter(RouterChan)
	for PD := range GlobalResvChan {
		var network bytes.Buffer        // Stand-in for a "network" connection
		dec := gob.NewDecoder(&network) // Will read from "network".

		network.Write(PD)

		InboundPacket := PeerPacket{}
		err := dec.Decode(&InboundPacket)
		if err != nil {
			logger.Printf("Unable to decode packet")
			continue
		}

		debuglogger.Printf("New Packet:")
		debuglogger.Printf("New Packet: Host: %s", InboundPacket.Host)
		debuglogger.Printf("New Packet: Service: %s", InboundPacket.Service)
		debuglogger.Printf("New Packet: Message: %s", InboundPacket.Message)
		debuglogger.Printf("New Packet: Salt: %s", InboundPacket.Salt)

		if !SeenPacketBefore(InboundPacket) {
			SendPacket(InboundPacket)
			RouterChan <- InboundPacket
		}

	}
}

func PacketRouter(inbound chan PeerPacket) {
	for PD := range inbound {
		if PD.Service == "PEX_REQUEST" {
			MakePEXPacket()
		}
		if PD.Service == "PEX" {
			ProcessPEXPacket(PD)
		}
		if PD.Service == "" {

		}
		if ConnectedIRCClients.Clients != nil {
			for _, v := range ConnectedIRCClients.Clients {
				if v.Channels[PD.Service] {
					v.InboundChan <- PD
				}
			}
		}
	}
}

func SendPacket(P PeerPacket) {
	HN, e := os.Hostname()
	if e != nil {
		logger.Printf("Could not send packet, Could not read hostname for reason %s", e.Error())
		return
	}
	if P.Host == "" {
		P.Host = HN
	}
	if P.Salt == "" {
		P.Salt = fmt.Sprintf("%s%x", RandString(7), HashValue([]byte(P.Message))[:5])
	}
	SeenPacketBefore(P) // To put the seed in the system.

	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	e = enc.Encode(&P)

	if e != nil {
		logger.Printf("Could not encode a packet(!) Service: %s", P.Service)
	}

	GlobalPeerList.m.Lock()
	Dispatch := network.Bytes()
	debuglogger.Printf("Sending packet %x", HashValue(Dispatch))

	debuglogger.Printf("Dispatching packet to all nodes size: %d", len(Dispatch))
	for _, Host := range GlobalPeerList.Peers {
		debuglogger.Printf("~!~ Host: %s IsAlive: %s", Host.ApparentIP, Host.Alive)
		if Host.Alive {
			if Host.MessageChan != nil {
				Host.MessageChan <- Dispatch
			}
		}
	}
	GlobalPeerList.m.Unlock()
}

// Check to see if the packet has been already seen,
// if it has not then it will add it to the cache (evicting old stuff if needed)
func SeenPacketBefore(P PeerPacket) bool {
	for k, v := range PacketCache {
		if v != nil {
			if v.Salt == P.Salt {
				PacketCache[k].LastSeen = time.Now().Unix()
				return true
			}
		}
	}
	// Add it to the cache
	MaxItemsInCache := 1000
	LowestTime := -1 // Max (int64) -1
	OldestItem := -1

	for i := 0; i < MaxItemsInCache; i++ {
		if PacketCache[i] != nil {
			if PacketCache[i].LastSeen < int64(LowestTime) || LowestTime == -1 {
				OldestItem = i
				if PacketCache[i].LastSeen == 0 {
					// Its a free slot!
					break
				}
				LowestTime = int(PacketCache[i].LastSeen)
			}
		} else {
			OldestItem = i
			break
		}
	}

	PacketCache[OldestItem] = &P
	PacketCache[OldestItem].LastSeen = time.Now().Unix()
	return false
}
