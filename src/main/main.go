package main

import (
//	"net"
	"../network/bcast"
	"../network/localip"
	// ".project-gruppa/network/bcast"
	// ".project-gruppa/network/peers"
	// "flag"
	"fmt"
	"os"
	"time"
	// "time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	UDP_broadcast_IP := make(chan string)
	IP_recieved := make(chan string)
	send_port := 20012
	read_port := 20013

	var id string
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
		fmt.Println("My IP: ", localIP)
		go bcast.Transmitter(send_port, UDP_broadcast_IP)
		go bcast.Receiver(read_port, IP_recieved)
		for{
			UDP_broadcast_IP <- localIP
			time.Sleep(time.Second)
			fmt.Println("Sent")
			fmt.Println("Recieved: ", <- IP_recieved)
		}
		}
}
