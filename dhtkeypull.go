package main

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
)

func PullDHTKey() {
	logger.Printf("Bootstrapping to find peers who have keys...")

}

func AttemptToPullKeyFromHost(host string) error {
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
		return err
	}
	defer client.Close()
	keyses, requests, err := client.OpenChannel("keys", nil)
	if err != nil {
		return err
	}
	go ssh.DiscardRequests(requests)

	CPrivKey := make([]byte, 25565)

	read, err := keyses.Read(CPrivKey)

	if err != nil {
		return err
	}

	PrivKey := DecryptText(CPrivKey[:read], []byte(CC_KEY))
	if len(PrivKey) == 0 {
		return errors.New("Could not decrypt key. Probs a mismatch in server keys.")
	}

	return nil
}
