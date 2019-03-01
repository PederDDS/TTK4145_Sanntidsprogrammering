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

const(
  LISTEN_PORT = ":20012"
  LISTEN_TO_SERVER = ":30000"
)

func main() {
	listenAddr, err := net.ResolveUDPAddr("udp",LISTEN_PORT)
	checkError(err)

	listenConn, err := net.ListenUDP("udp", listenAddr)
	checkError(err)
	buf := make([]byte, 1024)

  for{
	length, addr, err := listenConn.ReadFromUDP(buf)
  checkError(err)
  fmt.Println("Received ", string(buf[0:length]), " from ", addr)
  }
}
