package main

import (
	"fmt"
	"net"
	"time"
)


func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}


func main() {
  // Defining IP-adresses and ports
  SERVER_IP := "10.100.23.242"
  SERVER_PORT := ":33546"
  LOCAL_IP  := "10.100.XXX.XXX"
  LOCAL_PORT  := ":XXXX"


	server_addr, err := net.ResolveTCPAddr("tcp", SERVER_IP + SERVER_PORT)
	checkError(err)

	server_conn, err := net.DialTCP("tcp", nil, server_addr)
	checkError(err)

  // Make the server connect to you at LOCAL_PORT
	connect_order := "Connect to: " + LOCAL_IP + LOCAL_PORT + "\x00"


	local_addr, err := net.ResolveTCPAddr("tcp", LOCAL_IP + LOCAL_PORT)
	checkError(err)

	listener, err := net.ListenTCP("tcp", local_addr)
	checkError(err)

	// Send the connection order
	_, err = server_conn.Write([]byte(connect_order))
	checkError(err)

  // Establish TCP connection
	client_conn, err := listener.AcceptTCP()
	checkError(err)

	for {

		msg := "This is a message from XXXX, \x00"
		fmt.Println("Sending message: ", msg, "\n")
		client_conn.Write([]byte(msg))

		buffer := make([]byte, 1024)
		client_conn.Read(buffer)
		fmt.Println("Recived message: ", string(buffer), "\n")

		time.Sleep(1 * time.Second)
	}

}
