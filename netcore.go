package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"io/ioutil"
	"net"
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

func HandleIncomingConn(nConn net.Conn, config *ssh.ServerConfig, IsUserAllowedKeyAuth map[string]bool) {
	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	defer nConn.Close()
	if err != nil {
		debuglogger.Printf("WARNING - Was unable to handshake with %s RSN %s", nConn.RemoteAddr().String(), err)
		return
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "keys" && newChannel.ChannelType() != "northstar" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			debuglogger.Printf("WARNING - Rejecting %s Because they asked for a chan type %s that I don't have", nConn.RemoteAddr().String(), newChannel.ChannelType())
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			debuglogger.Printf("WARNING - Was unable to Accept channel with %s", nConn.RemoteAddr().String())
			return
		}
		go ssh.DiscardRequests(requests)

		if newChannel.ChannelType() == "keys" {
			// Send them awayyyy
			channel.Write(EncryptText(PEM_KEY, []byte(CC_KEY)))
		} else if newChannel.ChannelType() == "northstar" {
			if IsUserAllowedKeyAuth[nConn.RemoteAddr().String()] {

				go HandleNorthStarChan(channel, nConn)
				AskForPEX()
			} else {
				logger.Printf("Non key authed user tried to use NS channel (Attempted attack?) [%s]", nConn.RemoteAddr().String())
				nConn.Close() // Go away, Stop trying to be a faaake
			}
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
	Chan, requests, err := client.OpenChannel("northstar", nil)
	if err != nil {
		client.Close()
		return err
	}
	go ssh.DiscardRequests(requests)
	WriteChan := make(chan []byte)
	P.Alive = true
	P.MessageChan = WriteChan

	go NSConnWriteDrain(WriteChan, Chan, P)
	go NSConnReadDrain(GlobalResvChan, Chan, P)
	return nil
}
