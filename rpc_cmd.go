package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunRPCShell(bashline string, jobid int) {
	out, err := exec.Command("sh", "-c", bashline).Output()
	Pack := PeerPacket{}
	Pack.Service = "RPC"
	if err != nil {
		Pack.Message = fmt.Sprintf("~%d!! Failed to run command - %s", jobid, err.Error())
		SendPacket(Pack)
	}

	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		Pack.Message = fmt.Sprintf("~%d [%d/%d] Failed to run command - %s", jobid, i, len(lines), lines[i])
		SendPacket(Pack)
	}
}
