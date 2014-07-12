package main

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"time"
)

func PullDHTKey() []byte {
	logger.Printf("Bootstrapping to find peers who have keys...")
	time.Sleep(time.Second * 5) // Give the DHT some time to get hosts
	for {
		//if GlobalPeerList.PeerCount != 0 {
		for k, v := range GlobalPeerList.Peers {
			debuglogger.Printf("DEBUG: Looking in the Peer list, Going to try and pull key from %s %d", v, k)
			err, out := AttemptToPullKeyFromHost(v.ApparentIP)
			if err != nil {
				logger.Printf("Failed to pull key from: %s RSN %S\n", v.ApparentIP, err)
			} else {
				logger.Printf("Pulled key from: %s\n", v.ApparentIP)
				return out
			}
		}
		//}
		time.Sleep(time.Second * 5)
	}
}

func AttemptToPullKeyFromHost(host string) (e error, out []byte) {
	config := &ssh.ClientConfig{
		User: "gimmekeys",
		Auth: []ssh.AuthMethod{
			// ClientAuthPassword wraps a ClientPassword implementation
			// in a type that implements ClientAuth.
			ssh.Password("gimmekeys"),
		},
	}
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return err, out
	}
	defer client.Close()
	keyses, requests, err := client.OpenChannel("keys", nil)
	if err != nil {
		return err, out
	}
	go ssh.DiscardRequests(requests)

	CPrivKey := make([]byte, 25565)

	read, err := keyses.Read(CPrivKey)

	if err != nil {
		return err, out
	}

	PrivKey := DecryptText(CPrivKey[:read], []byte(CC_KEY))
	if len(PrivKey) == 0 {
		return errors.New("Could not decrypt key. Probs a mismatch in server keys."), out
	}

	return nil, PrivKey
}
