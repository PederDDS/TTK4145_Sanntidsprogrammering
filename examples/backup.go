package main

import (
  "fmt"
  "time"
  "net"
  "encoding/json"
  "log"
  "os/exec"
)


func primary(counter int, UDPBroadcast *net.UDPConn) {
  //create new backup
  //er ikke 100% sikker på inputargumentene til Command http://manpages.ubuntu.com/manpages/xenial/man1/gnome-terminal.1.html
  //er derimot 100% sikker på at denne måten å gjøre det på bare er ubuntu-vennlig, må sikkert ha med noe med gnome

  // På Ubuntu:
  //backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go") //https://stackoverflow.com/questions/26416887/golang-opening-second-terminal-console

  // På Windows
  backup := exec.Command("cmd", "main", "go run backup.go")
  err := backup.Run()
  checkError(err)

  for n := counter; ; n++ {
    buf, _ := json.Marshal(n)
    UDPBroadcast.Write(buf)
    time.Sleep(200*time.Millisecond)
  }
}



func backup(listenConn *net.UDPConn) int {
  listenCh := make(chan int,1)
  backupCounter := 0
  go listen(listenCh, listenConn)

  for {
    select {
      case backupCounter = <- listenCh:
        time.Sleep(100*time.Millisecond)
      case <- time.After(1*time.Second):
        return backupCounter
      }
    }
}

func listen(listenCh chan int, listenConn *net.UDPConn) {
  buf := make([]byte, 64)
  counter := 0

  for {
    _, _, err := listenConn.ReadFromUDP(buf[:])
    checkError(err)
    json.Unmarshal(buf, counter)
    fmt.Println(counter)
    listenCh <- counter
    time.Sleep(100*time.Millisecond)
  }
}


func checkError(err error)  {
  if err!=nil {
    fmt.Println("Error: ", err)
    log.Fatal(err)
  }
}


func main() {
  udpAddr, err := net.ResolveUDPAddr("udp", ":20012")
  checkError(err)

  listen, err := net.ListenUDP("udp", udpAddr)
  checkError(err)

  backupVal := backup(listen)
  listen.Close()

  udpAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
  checkError(err)

  udpBroadcast, err := net.DialUDP("udp", nil, udpAddr2)
  checkError(err)

  primary(backupVal, udpBroadcast)

  udpBroadcast.Close()
}
