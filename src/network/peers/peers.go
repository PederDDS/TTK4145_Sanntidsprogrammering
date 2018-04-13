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

	go Transmitter(def.SEND_ID_PORT, string(def.LOCAL_ID), transmitEnable)
	go PollNetwork(peerUpdateCh)

	currentMap := ordermanager.GetElevMap()

	for {
		select {
		case msg := <- peerUpdateCh:
			for element := range len(msg.Lost) {
				id := int(msg.Lost[element])
				currentMap[id].State = def.S_Dead
			}
			if msg.New != "" {
				id := int(msg.New)
				currentMap[id].State = def.S_Init
			}

			}

		}
	sendMsg := def.MakeMapMessage(currentMap, nil)
	msg_deadElev <- sendMsg
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
				fmt.Println("Something came from the network")
				peerUpdateCh <- msg_fromNet
		}
	}
}
