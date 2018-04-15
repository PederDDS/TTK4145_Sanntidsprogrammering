package peers

import (
	"../conn"
	"fmt"
	"net"
	"sort"
	"time"
	"../../ordermanager"
	"../../def"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 50 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])


		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}

func PeerWatch(msg_deadElev chan<- def.MapMessage)  {
	transmitEnable 	:= make(chan bool, 100)
	peerUpdateCh		:= make(chan PeerUpdate, 100)

	var sendID string
	var ID int

	switch def.LOCAL_ID {
	case 0:
		sendID = "ljhvcada"
	case 1:
		sendID = "esoiufhwep"
	case 2:
		sendID = "adwjpafae"
	case 3:
		sendID = "sdlhifsake"
	}

	go Transmitter(def.SEND_ID_PORT, sendID, transmitEnable)
	go PollNetwork(peerUpdateCh)


	var currentMap ordermanager.ElevatorMap
	var send bool

	for {
		send = false
		select {
		case msg := <- peerUpdateCh:
			currentMap = ordermanager.GetElevMap()
			 if msg.New != ""{
				switch msg.New {
				case "ljhvcada":
					fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
					ID = 0
					send = true
				case "esoiufhwep":
					fmt.Println("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
					ID = 1
					send = true
				case "adwjpafae":
					fmt.Println("cccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
					ID = 2
					send = true
				case "sdlhifsake":
					fmt.Println("dddddddddddddddddddddddddddddddddddddddddddddddddddddd")
					ID = 3
					send = true
				}

				if send {
					fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
					currentMap[ID].State = def.S_Idle

					sendMsg := def.MakeMapMessage(currentMap, nil)
					msg_deadElev <- sendMsg
			}

			} else if len(msg.Lost) > 0 {
				if msg.Lost[0] != ""{
   				switch msg.Lost[0] {
   				case "ljhvcada":
					fmt.Println("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
   					ID = 0
						send = true
   				case "esoiufhwep":
					fmt.Println("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
   					ID = 1
						send = true
   				case "adwjpafae":
					fmt.Println("gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg")
   					ID = 2
						send = true
   				case "sdlhifsake":
					fmt.Println("hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh")
   					ID = 3
						send = true
   				}
					if send {
						fmt.Println("ÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆÆ")
						currentMap[ID].State = def.S_Dead

						sendMsg := def.MakeMapMessage(currentMap, nil)
						msg_deadElev <- sendMsg
				}
			}
		}

		}
	}
}

func PollNetwork(peerUpdateCh chan<- PeerUpdate){
	poll_chn := make(chan PeerUpdate, 100)
	for port := 20010; port < 20100; port++{
		if port != def.SEND_ID_PORT{
			go Receiver(port, poll_chn)
		}
	}
	for {
			select {
			case msg_fromNet := <- poll_chn:
				peerUpdateCh <- msg_fromNet
		}
	}
}
