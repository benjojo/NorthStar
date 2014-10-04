package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"io/ioutil"
	"net"
	"time"
)

// 48563

func WaitForConnections() {
	publicpart := GSigner.PublicKey()
	IsUserAllowedKeyAuth := make(map[string]bool)

	// Setup logic for the SSH server.
	SSHConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			IsUserAllowedKeyAuth[conn.RemoteAddr().String()] = false
			if conn.User() == "gimmekeys" && string(pass) == "gimmekeys" {
				perms := ssh.Permissions{}
				logger.Println("Authed a Key Pull")
				return &perms, nil
			} else {
				return nil, errors.New("Auth Failed")
			}
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Compare(publicpart.Marshal(), key.Marshal()) == 0 {
				perms := ssh.Permissions{}
				IsUserAllowedKeyAuth[conn.RemoteAddr().String()] = true
				logger.Printf("Inbound Connection from %s", conn.RemoteAddr().String())
				return &perms, nil
			} else {
				IsUserAllowedKeyAuth[conn.RemoteAddr().String()] = false
				return nil, errors.New("Key does not match")
			}
		},
	}

	SSHConfig.AddHostKey(GSigner)
	listener, err := net.Listen("tcp", "0.0.0.0:48563")
	if err != nil {
		logger.Fatalln("Could not start TCP listening on 0.0.0.0:48563")
	}
	logger.Println("Waiting for TCP conns on 0.0.0.0:48563")

	for {
		nConn, err := listener.Accept()
		if err != nil {
			debuglogger.Println("WARNING - Failed to Accept TCP conn. RSN: %s / %s", err.Error(), err)
			continue
		}
		go HandleIncomingConn(nConn, SSHConfig, IsUserAllowedKeyAuth)
	}
}

func TimeoutConnection(Done chan bool, nConn net.Conn) {
	select {
	case <-Done:
		return
	case <-time.After(time.Second * 10):
		nConn.Close()
	}
}

func HandleIncomingConn(nConn net.Conn, config *ssh.ServerConfig, IsUserAllowedKeyAuth map[string]bool) {
	DoneCh := make(chan bool)
	go TimeoutConnection(DoneCh, nConn)
	sshconn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err == nil {
		DoneCh <- true
	}

	defer nConn.Close()
	if err != nil {
		debuglogger.Printf("WARNING - Was unable to handshake with %s RSN %s", nConn.RemoteAddr().String(), err)
		return
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "keys" && newChannel.ChannelType() != "northstar" && newChannel.ChannelType() != "nodeid" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			debuglogger.Printf("WARNING - Rejecting %s Because they asked for a chan type %s that I don't have", nConn.RemoteAddr().String(), newChannel.ChannelType())
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			debuglogger.Printf("WARNING - Was unable to Accept channel with %s", nConn.RemoteAddr().String())
			return // If we cannot accept the channel then we have probs ran out of ram, so lets disconnect this.
		}
		go ssh.DiscardRequests(requests)

		if newChannel.ChannelType() == "keys" {
			// Send them awayyyy
			channel.Write(EncryptText(PEM_KEY, []byte(CC_KEY)))
		} else if newChannel.ChannelType() == "northstar" {
			if IsUserAllowedKeyAuth[nConn.RemoteAddr().String()] {

				go HandleNorthStarChan(channel, nConn, sshconn)
				AskForPEX()
			} else {
				logger.Printf("Non key authed user tried to use NS channel (Attempted attack?) [%s]", nConn.RemoteAddr().String())
				nConn.Close() // Go away, Stop trying to be a faaake
			}
		} else if newChannel.ChannelType() == "nodeid" {
			channel.Write([]byte(NodeID))
			channel.Close()
		} else {
			logger.Printf("Unknown Channel Type, Dropping the connection to %s chan was %s", nConn.RemoteAddr().String(), newChannel.ChannelType())
			debuglogger.Printf("DEBUG: %x vs %x", newChannel.ChannelType(), "keys")
			return
		}
	}

}

func LoadPrivKeyFromFile(file string) []byte {
	privateBytes, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Fatalln("Failed to load private key")
	}
	return privateBytes
}

var GSigner ssh.Signer

func ConnectToPeer(P *Peer) error {

	config := &ssh.ClientConfig{
		User: "northstar",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(GSigner),
		},
	}
	client, err := ssh.Dial("tcp", P.ApparentIP, config)
	if err != nil {
		return err
	}
	IDChan, _, err := client.OpenChannel("nodeid", nil)
	if err != nil {
		client.Close()
		return err
	}
	IDBuffer := make([]byte, 22)
	in, err := IDChan.Read(IDBuffer)
	if err != nil {
		client.Close()
		return err
	}
	RemoteID := string(IDBuffer[:in])
	if RemoteID == NodeID {
		// Oh. Thats us.
		// Huh.
		GlobalPeerList.RemoveByStruct(*P)
		return err
	}

	// Good idea to check that there isnt any other connections with this ID.
	// If there is. Bash'em and let this one replace them.

	for _, v := range GlobalPeerList.Peers {
		if v.NodeID == RemoteID {
			v.Alive = false
			v.Conn.Close()
			logger.Printf("Dropping dupe connection to avoid loops from %s", v.ApparentIP)
		}
	}

	Chan, requests, err := client.OpenChannel("northstar", nil)
	if err != nil {
		client.Close()
		return err
	}
	go ssh.DiscardRequests(requests)
	WriteChan := make(chan []byte)
	P.Alive = true
	P.MessageChan = WriteChan
	P.Conn = client.Conn

	go NSConnWriteDrain(WriteChan, Chan, P)
	go NSConnReadDrain(GlobalResvChan, Chan, P)
	return nil
}
