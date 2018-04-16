package main

import (
	"fmt"
	"time"

	"../IO"
	"../def"
	"../fsm"
	"../network/bcast"
	"../network/peers"
	"../ordermanager"
)

func main() {

	IO.Init("localhost:15657", def.NUMFLOORS)
	ordermanager.InitElevMap()

	var motor_direction IO.MotorDirection

	msg_toNetwork := make(chan ordermanager.ElevatorMap, 100)
	msg_fromNetwork := make(chan ordermanager.ElevatorMap, 100)
	msg_fromFSM := make(chan def.MapMessage, 100)
	msg_deadElev := make(chan def.MapMessage, 100)

	drv_buttons := make(chan IO.ButtonEvent)
	drv_floors := make(chan int)
	fsm_chn := make(chan bool, 1)
	elevator_map_chn := make(chan def.MapMessage)

	go IO.PollButtons(drv_buttons)
	go IO.PollFloorSensor(drv_floors)

	motor_direction = IO.MD_Down

	go fsm.FSM(drv_buttons, drv_floors, fsm_chn, elevator_map_chn, motor_direction, msg_fromFSM, msg_deadElev)
	go bcast.Transmitter(def.SEND_MAP_PORT, msg_toNetwork)
	go bcast.PollNetwork(msg_fromNetwork)
	go peers.PeerWatch(msg_deadElev)

	transmitTicker := time.NewTicker(100 * time.Millisecond)
	var newMsg ordermanager.ElevatorMap
	transmitFlag := false

	for {
		select {

		case msg := <-msg_fromNetwork:
			fmt.Println("case msg_fromNetwork in main")
			newMap, changeMade := ordermanager.UpdateElevMap(msg)
			if changeMade {
				newMsg = newMap
				transmitFlag = true
			}

		case msg := <-msg_fromFSM:
			fmt.Println("case msg_fromFSM in main")
			recievedMap := msg.SendMap.(ordermanager.ElevatorMap)
			currentMap, changeMade := ordermanager.UpdateElevMap(recievedMap)

			newMsg = currentMap
			fsm.SetButtonLights(currentMap)

			if changeMade {
				transmitFlag = true
			}

		case <-transmitTicker.C:
			if transmitFlag {
				msg_toNetwork <- newMsg
				transmitFlag = false
			}
		default:
		}
	}
}
