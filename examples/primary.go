package main

import (
  "fmt"
  "time"
  "net"
  "encoding/json"
  "log"
)

func primary(counter int) {
    sendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
    checkError(err)

    conn, err := net.DialUDP("udp", nil, sendAddr)
    checkError(err)

    defer conn.Close()

    for n := counter; ; n++ {
      buf, _ := json.Marshal(n)
      conn.Write(buf)
      fmt.Println(n)
      time.Sleep(200*time.Millisecond)
    }
}

func checkError(err error){
    if err!=nil {
        log.Fatal(err)
        fmt.Println("Error message: ", err)
    }
}



func main() {
    primary(0)
}
