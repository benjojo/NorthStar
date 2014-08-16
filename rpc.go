package main

import (
	"code.google.com/p/go.crypto/ssh"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var UnlockTimeout int64

// Returns true if the system is in super user unlock mode.
func GetUnlockedState() bool {
	if time.Now().Unix() > UnlockTimeout {
		return false
	} else {
		return true
	}
}

// Evaluates if a unlock code is correct and will unlock the system if correct
func AttemptToUnlock(code string) error {
	if UnlockPub == nil {
		return fmt.Errorf("Unlocking is disabled due to lack of pub key")
	}

	// A unlock code is made up of 3 parts
	// <unix time that is in the last 1 min> <time in seconds to remain unlocked> <sig from key to verify operator>
	UnlockParts := strings.Split(code, " ")
	if len(UnlockParts) != 3 {
		return fmt.Errorf("Code Syntax is wrong.")
	}

	// Check to see if the first 2 parts are numbers
	SigTime, err := strconv.ParseInt(UnlockParts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Code Syntax is wrong. (invalid SigTime)")
	}

	UnlockWindow, err := strconv.ParseInt(UnlockParts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("Code Syntax is wrong. (invalid UnlockWindow)")
	}

	// Before checking if the sig is valid, lets check the easy stuff first.

	if math.Abs(float64(time.Now().Unix()-SigTime)) > 60 {
		return fmt.Errorf("Time sig is out of scope (Maybe system clock is incorrect)")
	}

	if UnlockWindow > 1800 {
		return fmt.Errorf("You cannot unlock the system for more than 30 mins")
	}

	sig, err := hex.DecodeString(UnlockParts[2])
	if err != nil {
		return fmt.Errorf("Code Syntax is wrong. (unable to de-hex the sig)")
	}
	sigstruct := ssh.Signature{}
	sigstruct.Format = "ssh-rsa"
	sigstruct.Blob = sig

	err = UnlockPub.Verify([]byte(fmt.Sprintf("%s %s", UnlockParts[0], UnlockParts[1])), &sigstruct)
	if err != nil {
		return fmt.Errorf("Sig failed to validate")
	}

	UnlockTimeout = time.Now().Unix() + UnlockWindow
	return nil
}

func ProcessRPCPacket(inbound PeerPacket) {
	if strings.HasPrefix(inbound.Message, "#") {
		// Operator wants to run a shell cmd
		Outbound := PeerPacket{
			Service: "RPC",
		}
		if !GetUnlockedState() {
			Outbound.Message = "System is not unlocked, Please Unlock to use this"
			SendPacket(Outbound)
			return
		}
		Outbound.Message = "Unsupported"
		SendPacket(Outbound)
	}

	if strings.HasPrefix(inbound.Message, "!") && len(inbound.Message) != 1 {
		// Operator wants to auth
		result := AttemptToUnlock(inbound.Message[1:])
		Outbound := PeerPacket{}
		Outbound.Service = "RPC"
		if result != nil {
			Outbound.Message = fmt.Sprintf("Unlock failed: %s", result.Error())
		} else {
			Outbound.Message = "System Unlocked."
		}
		SendPacket(Outbound)
	}

	if strings.HasPrefix(inbound.Message, "+") && len(inbound.Message) != 1 {
		// Operator wants to update NorthStar
		Outbound := PeerPacket{
			Service: "RPC",
		}
		if !GetUnlockedState() {
			Outbound.Message = "System is not unlocked, Please Unlock to use this"
			SendPacket(Outbound)
			return
		}
	}
}
