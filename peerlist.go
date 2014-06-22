package main

import (
	"net"
	"sync"
)

type Peer struct {
	HostName   string
	ApparentIP string
	Conn       net.Conn
	ID         int
	Alive      bool
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

func (p PList) RemoveByStruct(n Peer) {
	p.m.Lock()
	//p.Peers[n.ID].Alive = false
	if p.Peers[n.ID].Conn != nil {
		p.Peers[n.ID].Conn.Close()
	}
	p.m.Unlock()
}
