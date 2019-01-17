package server

import (
	"log"
	"os/exec"
)

func reboot() (err error) {
	if err = exec.Command("cmd", "/C", "shutdown", "/r", "/t", "1", "/f").Run(); err != nil {
		log.Printf("Failed to initiate reboot: %v", err)
	}
	return
}
