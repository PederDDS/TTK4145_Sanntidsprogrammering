package main

import (
	"fmt"
	"net"
)

func checkError(err error){
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

  return string(buf[25:length-1]) //Extracts only the server IP address
}

func main(){
	buffer 			:= make([]byte, 1024)
	LISTEN_PORT 		:= ":30000"
	COMM_PORT		:= ":200XX"

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
