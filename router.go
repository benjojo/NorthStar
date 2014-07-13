package main

import (
	"bytes"
	"encoding/gob"
	"time"
)

var GlobalResvChan chan []byte
var PacketCache map[int]*PeerPacket

func RelayPackets() {
	GlobalResvChan = make(chan []byte)
	PacketCache = make(map[int]*PeerPacket)

	for PD := range GlobalResvChan {
		var network bytes.Buffer        // Stand-in for a "network" connection
		dec := gob.NewDecoder(&network) // Will read from "network".

		InboundPacket := PeerPacket{}
		err := dec.Decode(&InboundPacket)
		if err != nil {
			logger.Printf("Unable to decode packet, Skipping")
			continue
		}

		if !SeenPacketBefore(InboundPacket) {
			GlobalPeerList.m.Lock()
			for _, Host := range GlobalPeerList.Peers {
				if Host.Alive {
					Host.MessageChan <- PD
				}
			}
			GlobalPeerList.m.Unlock()
		}

	}
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
	LowestTime := 9223372036854775806 // Max (int64) -1
	OldestItem := -1

	for i := 0; i < MaxItemsInCache; i++ {
		if PacketCache[i] != nil {
			if PacketCache[i].LastSeen < int64(LowestTime) {
				OldestItem = i
				if PacketCache[i].LastSeen == 0 {
					// Its a free slot!
					break
				}
				LowestTime = int(PacketCache[i].LastSeen)
			}
		}
	}

	PacketCache[OldestItem] = &P
	PacketCache[OldestItem].LastSeen = time.Now().Unix()
	return false
}
