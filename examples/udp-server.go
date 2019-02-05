package main

import (
	"fmt"
	"net"
)

func checkError(err error) bool {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func listenOnPort(port string) string {
	ServerAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	buf := make([]byte, 1024)

	length, _, err := ServerConn.ReadFromUDP(buf)
  checkError(err)

  return string(buf[0:n])
}

func main(){
	buffer 			:= make([]byte, 1024)
	LISTEN_PORT := ":30000"
	LOCAL_IP		:= "10.100.XXX.XXX"
	COMM_PORT		S:= ":200XX"

	SERVER_IP := listenOnPort(LISTEN_PORT)

	msg := "Hello from XX"

	localAddr, err := net.ResolveUDPAddr("udp", COMM_PORT)
  checkError(err)

	readConn, err := net.ListenUDP("udp", localAddr)
  checkError(err)

	sendConn, err := net.Dial("udp", SERVER_IP + COMM_PORT)
	checkError(err)

	sendConn.Write([]byte(msg))

	length, err := readConn.Read(buffer)
	checkError(err)

	fmt.Println(string(buffer[0:length]))

	sendConn.Close()
	readConn.Close()
}
