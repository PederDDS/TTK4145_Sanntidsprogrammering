package main

import (
	"net"
	"localip"
	// ".project-gruppa/network/bcast"
	// ".project-gruppa/network/peers"
	// "flag"
	// "fmt"
	"os"
	// "time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	var id string
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
}
