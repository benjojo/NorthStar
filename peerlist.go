package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
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

func (p *PList) Add(n *Peer) {
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
	RestorePeerList()

	hash := HashValue([]byte(CC_KEY))
	inboundchan := StartDHT(fmt.Sprintf("%x", hash[:20]))
	for host := range inboundchan {
		if !GlobalPeerList.ContainsIP(host) {
			NewPeer := Peer{}
			NewPeer.Alive = false
			NewPeer.ApparentIP = host
			GlobalPeerList.Add(&NewPeer)
			debuglogger.Printf("DEBUG: Added new peer to the peer list, Host is %s", host)
		}
	}
}

func AutoSavePeerList() {
	SaveList := ""
	for {
		SaveList = ""
		time.Sleep(time.Minute)
		GlobalPeerList.m.Lock()
		for _, v := range GlobalPeerList.Peers {
			if v.Alive {
				SaveList = SaveList + strings.Split(v.ApparentIP, ":")[0] + ":48563\n"
			}
		}
		GlobalPeerList.m.Unlock()
		err := ioutil.WriteFile("/.nspeerlistcache", []byte(SaveList), 660)

		if err != nil {
			debuglogger.Printf("Unable to save peer list to a cache because of %s", err)
		}
	}
}

func RestorePeerList() {
	b, err := ioutil.ReadFile("/.nspeerlistcache")
	if err != nil {
		logger.Printf("Cannot read peer list cache, not restoring from peer list")
		return
	}

	lines := strings.Split(string(b), "\n")
	for i := 0; i < len(lines); i++ {
		if !GlobalPeerList.ContainsIP(lines[i]) {
			NewPeer := Peer{}
			NewPeer.Alive = false
			NewPeer.ApparentIP = lines[i]
			GlobalPeerList.Add(&NewPeer)
		}
	}
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
