package main

import (
    "fmt"
    "net"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}


func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp",":30000")
	checkError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	buf := make([]byte, 1024)

	length, addr, err := ServerConn.ReadFromUDP(buf)
  checkError(err)
  fmt.Println("Received ",string(buf[0:length]), " from ",addr)
}
