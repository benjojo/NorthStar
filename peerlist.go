package main

import (
	"fmt"
	"net"
	"sync"
)

type Peer struct {
	HostName   string
	ApparentIP string
	Conn       net.Conn
	ID         int
	Alive      bool
	LastSeen   int64
}

type PList struct {
	Peers     map[int]Peer
	PeerCount int
	m         sync.Mutex
}

func (p PList) Add(n Peer) {
	p.m.Lock()
	p.PeerCount++
	n.ID = p.PeerCount
	p.Peers[p.PeerCount] = n
	p.m.Unlock()
}

func (p PList) ContainsIP(host string) bool {
	p.m.Lock()
	defer p.m.Unlock()
	for _, v := range p.Peers {
		if v.ApparentIP == host {
			logger.Printf("DEBUG %s LOOKS ALOT LIKE %s", v.ApparentIP, host)
			return true
		}
	}
	return false
}

func (p PList) RemoveByStruct(n Peer) {
	p.m.Lock()
	//p.Peers[n.ID].Alive = false
	if p.Peers[n.ID].Conn != nil {
		p.Peers[n.ID].Conn.Close()
	}
	p.m.Unlock()
}

var GlobalPeerList PList

func StartLookingForPeers() {
	GlobalPeerList = PList{}
	GlobalPeerList.Peers = make(map[int]Peer)
	GlobalPeerList.PeerCount = 0

	hash := HashValue([]byte(CC_KEY))
	inboundchan := StartDHT(fmt.Sprintf("%x", hash[:20]))
	for host := range inboundchan {
		if !GlobalPeerList.ContainsIP(host) {
			NewPeer := Peer{}
			NewPeer.Alive = false
			NewPeer.ApparentIP = host
			GlobalPeerList.Add(NewPeer)
		}
	}
}
