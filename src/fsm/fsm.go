package fsm

import (
	"fmt"
	"time"
	//"../network/bcast"
	"../IO"
	"../def"
	"../ordermanager"
)

var elevator_state def.ElevState = def.S_Dead
var motor_direction IO.MotorDirection
var initialized bool = false

func timer(timeout chan<- bool) {
	fmt.Println("Timer started")
	time.Sleep(5 * time.Second)
	timeout <- true
	fmt.Println("Timeout sent")
}

func Initialize(floor_detection <-chan int, fsm_chn chan<- bool, elevator_map_chn chan<- def.MapMessage, direction IO.MotorDirection) {
	currentMap := ordermanager.GetElevMap()

	timeout := make(chan bool, 1)
	elevator_state = def.S_Init
	currentMap[def.LOCAL_ID].State = elevator_state
	motor_direction = direction
	currentMap[def.LOCAL_ID].Dir = motor_direction
	sendMessage := def.MakeMapMessage(currentMap, nil)
	fmt.Println("updateElevMap Initialize")
	newMap, _ := ordermanager.UpdateElevMap(sendMessage.SendMap.(ordermanager.ElevatorMap))
	IO.SetMotorDirection(motor_direction)
	go timer(timeout)
	newMap[1].Dir = 0 // fordi newMap is declared and not used...

	select {
	case floor := <-floor_detection:
		fmt.Println("Floor detected")
		motor_direction = IO.MD_Stop
		sendMessage = def.MakeMapMessage(nil, motor_direction)
		currentMap[def.LOCAL_ID].Dir = motor_direction
		sendMessage := def.MakeMapMessage(currentMap, nil)

		newMap, _ := ordermanager.UpdateElevMap(sendMessage.SendMap.(ordermanager.ElevatorMap))
		newMap[1].Dir = 0 // fordi newMap is declared and not used...
		IO.SetMotorDirection(motor_direction)
		elevator_state = def.S_Idle
		currentMap[def.LOCAL_ID].State = elevator_state
		currentMap[def.LOCAL_ID].Floor = floor
		sendMessage = def.MakeMapMessage(currentMap, nil)

		newMap, _ = ordermanager.UpdateElevMap(sendMessage.SendMap.(ordermanager.ElevatorMap))
		time.Sleep(2 * time.Second)
		fsm_chn <- true
	case <-timeout:
		fmt.Println("Timed out")
		Initialize(floor_detection, fsm_chn, elevator_map_chn, -motor_direction)
	}
}

func FSM(drv_buttons chan IO.ButtonEvent, drv_floors chan int, fsm_chn chan bool, elevator_map_chn chan def.MapMessage, direction IO.MotorDirection, msg_fromFSM chan def.MapMessage, msg_deadElev chan def.MapMessage) {

	if initialized == false {
		Initialize(drv_floors, fsm_chn, elevator_map_chn, IO.MD_Up)
		<-fsm_chn
		initialized = true
		fmt.Println("Elevator is initialized!")
	}

	doorTimer := time.NewTimer(def.DOOR_TIMEOUT_TIME * time.Second)
	doorTimer.Stop()
	idleTimer := time.NewTimer(def.IDLE_TIMEOUT_TIME * time.Second)

	for {
		select {
		case <-fsm_chn:
			currentMap:=ordermanager.GetElevMap()
			motor_direction = ChooseDirection(currentMap)
			if motor_direction != IO.MD_Stop {
				currentMap[def.LOCAL_ID].State = def.S_Moving
				fmt.Println("case fsm_chn in FSM")
				elevator_state = def.S_Moving
			} else if motor_direction == IO.MD_Stop {
				currentMap[def.LOCAL_ID].State = def.S_Idle
				elevator_state = def.S_Idle
			}
			SendMapMessage(msg_fromFSM, currentMap, nil)

		case arrivalFloor := <-drv_floors: //elevator arrives at new floor
			fmt.Println("case msg from drv_floors FSM")
			//currentMap := ordermanager.GetElevMap()

			FloorArrival(msg_fromFSM, arrivalFloor, doorTimer)
			idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)

/*
			currentMap = ordermanager.GetElevMap()
			motor_direction = ChooseDirection(currentMap)

			//update currentMap
			if motor_direction != IO.MD_Stop && currentMap[def.LOCAL_ID].State != def.S_Dead {
				currentMap[def.LOCAL_ID].State = def.S_Moving
				fmt.Println("case arrivalFloor in FSM")
				elevator_state = def.S_Moving
			}
			currentMap[def.LOCAL_ID].Dir = motor_direction
			currentMap[def.LOCAL_ID].Floor = arrivalFloor
			currentMap[def.LOCAL_ID].State = elevator_state

			//send map to main
			SendMapMessage(msg_fromFSM, currentMap, nil)
*/
		case msg := <-msg_deadElev: //elevator is dead
			fmt.Println("case message from msg_deadElev in FSM")
			localMap := ordermanager.GetElevMap()
			for elev := 0; elev < def.NUMELEVATORS; elev++ {
				if msg.SendMap.(ordermanager.ElevatorMap)[elev].State == def.S_Dead && localMap[elev].State != def.S_Dead && msg.SendEvent == "Dead elevator"{
					fmt.Println("FSM thinks that the dead elev is:", elev)
					DeadElevator(msg_fromFSM, elev)
					idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)
					if msg.SendMap.(ordermanager.ElevatorMap)[def.LOCAL_ID].State == def.S_Dead {
						elevator_state = def.S_Dead
					}
				} else if msg.SendMap.(ordermanager.ElevatorMap)[elev].State == def.S_Idle && localMap[elev].State == def.S_Dead && msg.SendEvent == "New elevator"{
					localMap := ordermanager.GetElevMap()
					localMap = ordermanager.OverWriteIdle(localMap, elev)
					if msg.SendMap.(ordermanager.ElevatorMap)[def.LOCAL_ID].State != def.S_Dead {
						elevator_state = def.S_Idle
					}

					SendMapMessage(msg_fromFSM, localMap, nil)
				}
			}

		case msg_button := <-drv_buttons: //detects new buttons pushed
			fmt.Println("case msg from drv_buttons in FSM")
			currentMap := ordermanager.GetElevMap()
/*
			ButtonPushed(msg_fromFSM, msg_button.Floor, int(msg_button.Button), doorTimer, currentMap)
			idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)
*/

			if msg_button.Button == IO.BT_Cab {
				fmt.Println("Cab order detected")

				currentMap[def.LOCAL_ID].Orders[msg_button.Floor][msg_button.Button] = ordermanager.ORDER

				currentMap, _ = ordermanager.UpdateElevMap(currentMap)
				ButtonPushed(msg_fromFSM, msg_button.Floor, int(msg_button.Button), doorTimer, currentMap)
				idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)

			} else if msg_button.Button == IO.BT_HallUp || msg_button.Button == IO.BT_HallDown {
				fmt.Println("Hall order detected")
				for elev := 0; elev < def.NUMELEVATORS; elev++ {
					if currentMap[elev].State != def.S_Dead && currentMap[elev].State != def.S_Init {
						currentMap[elev].Orders[msg_button.Floor][msg_button.Button] = ordermanager.ORDER
					}
				}
				currentMap, _ = ordermanager.UpdateElevMap(currentMap)
				ButtonPushed(msg_fromFSM, msg_button.Floor, int(msg_button.Button), doorTimer, currentMap)
				idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)
			}


			default:
			}
		}
}


func ChooseDirection(currentMap ordermanager.ElevatorMap) IO.MotorDirection {
	currentFloor := currentMap[def.LOCAL_ID].Floor

	if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_Cab] == ordermanager.ORDER_ACCEPTED {
		return IO.MD_Stop
	}

	switch currentMap[def.LOCAL_ID].Dir {
	case IO.MD_Up:
		fmt.Println("Choose Direction, moving up")
		if currentMap[def.LOCAL_ID].Floor != def.NUMFLOORS -1 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] == ordermanager.ORDER_ACCEPTED {
				fmt.Println("Jeg kjører opp og det er en ordre i riktig retning, så jeg stopper her")
				return IO.MD_Stop
			}

			if OrderAbove(currentMap) {
				fmt.Println("Jeg kom til en etasje og det er en ordre over, så jeg fortsetter")
				return IO.MD_Up
			}
		}

		if currentMap[def.LOCAL_ID].Floor != 0 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] == ordermanager.ORDER_ACCEPTED /* && !OrderBelow(currentMap)*/{
				fmt.Println("Jeg kom til en etasje og det er ingen ordre over, men jeg stopper her fordi noen skal ned")
				return IO.MD_Stop
			}

			if OrderBelow(currentMap) {
				fmt.Println("Jeg snur retning fordi det ikke er noe over meg, nen under")
				return IO.MD_Down
			}
		}

		if currentMap[def.LOCAL_ID].Floor == def.NUMFLOORS -1 {
			return IO.MD_Stop
		}

		return IO.MD_Stop

	case IO.MD_Down:
		fmt.Println("Choose Direction, moving down")
		if currentMap[def.LOCAL_ID].Floor != 0 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] == ordermanager.ORDER_ACCEPTED {
				fmt.Println("Jeg kjører ned og det er en ordre i riktig retning, så jeg stopper her")
				return IO.MD_Stop
			}

			if OrderBelow(currentMap) {
				fmt.Println("Jeg kom til en etasje og det er en ordre under, så jeg fortsetter")
				return IO.MD_Down
			}
		}

		if currentMap[def.LOCAL_ID].Floor != def.NUMFLOORS -1 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] == ordermanager.ORDER_ACCEPTED /* && !OrderAbove(currentMap)*/{
				fmt.Println("Jeg kom til en etasje og det er ingen ordre under, men jeg stopper her fordi noen skal opp")
				return IO.MD_Stop
			}

			if OrderAbove(currentMap) {
				fmt.Println("Jeg snur retning fordi det ikke er noe under meg, nen over")
				return IO.MD_Up
			}
		}

		if currentMap[def.LOCAL_ID].Floor == 0 {
			return IO.MD_Stop
		}

		return IO.MD_Stop

	case IO.MD_Stop:
		if currentMap[def.LOCAL_ID].Floor != ordermanager.NO_ORDER {
			for floor := currentMap[def.LOCAL_ID].Floor - 1; floor > -1; floor-- {
				if ordermanager.IsOrderOnFloor(currentMap, floor) || ordermanager.IsClosestElevator(currentMap, floor) {
					return IO.MD_Down
				}
			}
		}

		if currentMap[def.LOCAL_ID].Floor != def.NUMFLOORS {
			for floor := currentMap[def.LOCAL_ID].Floor + 1; floor < def.NUMFLOORS; floor++ {
				if ordermanager.IsOrderOnFloor(currentMap, floor) || ordermanager.IsClosestElevator(currentMap, floor) {
					return IO.MD_Up
				}
			}
		}

		return IO.MD_Stop

	default:
		return IO.MD_Stop
	}
}

func OrderAbove(currentMap ordermanager.ElevatorMap) bool {
	for floor := currentMap[def.LOCAL_ID].Floor + 1; floor < def.NUMFLOORS; floor++ {
		if ordermanager.IsOrderOnFloor(currentMap, floor) {
			return true
		}
	}
	return false
}

func OrderBelow(currentMap ordermanager.ElevatorMap) bool {
	for floor := 0; floor < currentMap[def.LOCAL_ID].Floor; floor++ {
		if ordermanager.IsOrderOnFloor(currentMap, floor) {
			return true
		}
	}
	return false
}

func DeleteOrdersOnFloor(currentMap ordermanager.ElevatorMap, currentFloor int) ordermanager.ElevatorMap {
	//also turns off light for the orders deleted
	if currentMap[def.LOCAL_ID].Dir == IO.MD_Up || currentMap[def.LOCAL_ID].Dir == IO.MD_Stop {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] = ordermanager.NO_ORDER
	} else if currentMap[def.LOCAL_ID].Dir == IO.MD_Down || currentMap[def.LOCAL_ID].Dir == IO.MD_Stop {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] = ordermanager.NO_ORDER
	} else if currentFloor == 0 || currentFloor == def.NUMFLOORS - 1 {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] = ordermanager.NO_ORDER
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] = ordermanager.NO_ORDER
		}

	currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_Cab] = ordermanager.NO_ORDER
	ordermanager.SetElevMap(currentMap)
	//SetButtonLights(currentMap)
	return currentMap
}


func FloorArrival(msg_fromFSM chan def.MapMessage, arrivalFloor int, doorTimer *time.Timer) {
	currentMap := ordermanager.GetElevMap()
	//SendMapMessage(msg_fromFSM, currentMap, nil) // Kun for testing!!!!!!!!!!!
	currentMap[def.LOCAL_ID].Floor = arrivalFloor
	IO.SetFloorIndicator(arrivalFloor)

	//check if there is an order on floor, if so delete orders (also do lights)
	switch currentMap[def.LOCAL_ID].State {
	case def.S_Idle:
		fmt.Println("case def.S_Idle in FloorArrival")

		//if order on floor, delete orders and set door open
		if ordermanager.IsOrderOnFloor(currentMap, arrivalFloor) {
			//direction i stop and door is set open
			motor_direction = IO.MD_Stop
			elevator_state = def.S_DoorOpen

			IO.SetMotorDirection(motor_direction)
			IO.SetDoorOpenLamp(true)

			currentMap = DeleteOrdersOnFloor(currentMap, arrivalFloor)
			fmt.Println("DeleteOrdersOnFloor i case idle i FloorArrival")
			currentMap[def.LOCAL_ID].Dir = motor_direction
			currentMap[def.LOCAL_ID].State = elevator_state
			SendMapMessage(msg_fromFSM, currentMap, nil)

			//wait for door timer, then set state to idle
			doorTimer.Reset(def.DOOR_TIMEOUT_TIME * time.Second)
			<-doorTimer.C
			DoorTimeout(msg_fromFSM)
		}

		motor_direction = ChooseDirection(currentMap)
		if motor_direction != IO.MD_Stop {
			elevator_state = def.S_Moving
		} else {
			elevator_state = def.S_Idle
		}
		IO.SetMotorDirection(motor_direction)

		currentMap[def.LOCAL_ID].Dir = motor_direction
		currentMap[def.LOCAL_ID].State = elevator_state
		SendMapMessage(msg_fromFSM, currentMap, nil)

	case def.S_Moving:
		fmt.Println("case def.S_Moving in FloorArrival")
		motor_direction = ChooseDirection(currentMap)
		IO.SetMotorDirection(motor_direction)

		if motor_direction == IO.MD_Stop && ordermanager.IsOrderOnFloor(currentMap, arrivalFloor) {

			IO.SetMotorDirection(motor_direction)
			elevator_state = def.S_DoorOpen

			IO.SetDoorOpenLamp(true)
			doorTimer.Reset(def.DOOR_TIMEOUT_TIME * time.Second)
			fmt.Println(currentMap)
			currentMap := DeleteOrdersOnFloor(currentMap, arrivalFloor)
			fmt.Println("DeleteOrdersOnFloor i case moving i FloorArrival")
			currentMap[def.LOCAL_ID].State = elevator_state
			currentMap[def.LOCAL_ID].Dir = motor_direction
			SendMapMessage(msg_fromFSM, currentMap, nil)

			<-doorTimer.C
			DoorTimeout(msg_fromFSM)
		} else {
				if motor_direction != IO.MD_Stop {
					elevator_state = def.S_Moving
				} else {
					elevator_state = def.S_Idle
				}
				currentMap[def.LOCAL_ID].Dir = motor_direction
				currentMap[def.LOCAL_ID].State = elevator_state
				SendMapMessage(msg_fromFSM, currentMap, nil)

		}

	case def.S_DoorOpen:
		doorTimer.Reset(def.DOOR_TIMEOUT_TIME * time.Second)
		//SetButtonLights(currentMap)
		<-doorTimer.C
		DoorTimeout(msg_fromFSM)
	}
}

func DeadElevator(msg_fromFSM chan def.MapMessage, deadElevId int) {
			fmt.Println("func: DeadElevator(", deadElevId,")")
			currentMap := ordermanager.GetElevMap()
			currentMap = ordermanager.OverWriteDead(currentMap, deadElevId)
			currentMap = ordermanager.RedistributeOrders(currentMap, deadElevId)

			SendMapMessage(msg_fromFSM, currentMap, nil)

}



func ButtonPushed(msg_fromFSM chan def.MapMessage, floor int, button int, doorTimer *time.Timer, currentMap ordermanager.ElevatorMap) {
	fmt.Println("buttonPushed")
	SetButtonLights(currentMap)

	switch currentMap[def.LOCAL_ID].State {
	case def.S_Idle:
		fmt.Println("state in button pushed is idle")
		if currentMap[def.LOCAL_ID].Floor == floor {
			fmt.Println("current floor in case idle")
			motor_direction = IO.MD_Stop
			IO.SetMotorDirection(motor_direction)
			currentMap = DeleteOrdersOnFloor(currentMap, floor)
			fmt.Println("DeleteOrdersOnFloor i case idle i buttonPushed")

			IO.SetDoorOpenLamp(true)
			doorTimer.Reset(def.DOOR_TIMEOUT_TIME * time.Second)

			elevator_state = def.S_DoorOpen
			currentMap[def.LOCAL_ID].State = elevator_state
			currentMap[def.LOCAL_ID].Dir = motor_direction
			SendMapMessage(msg_fromFSM, currentMap, nil)

			<-doorTimer.C
			DoorTimeout(msg_fromFSM)

		} else {
			motor_direction = ChooseDirection(currentMap)
			fmt.Println("new motor direction is: ", motor_direction)
			IO.SetMotorDirection(motor_direction)

			if motor_direction != IO.MD_Stop {
				elevator_state = def.S_Moving
				fmt.Println("case def.S_Idle in button pushed")
			}

			currentMap[def.LOCAL_ID].Dir = motor_direction
			currentMap[def.LOCAL_ID].State = elevator_state
			SendMapMessage(msg_fromFSM, currentMap, nil)
		}

	case def.S_DoorOpen:
		if currentMap[def.LOCAL_ID].Floor == floor {
			currentMap = DeleteOrdersOnFloor(currentMap, floor)
			fmt.Println("DeleteOrdersOnFloor i case doorOpen i buttonPushed")
			doorTimer.Reset(def.DOOR_TIMEOUT_TIME * time.Second)

			SendMapMessage(msg_fromFSM, currentMap, nil)
			<-doorTimer.C
			DoorTimeout(msg_fromFSM)
		} else {
			if currentMap[def.LOCAL_ID].Orders[floor][button] == ordermanager.ORDER_ACCEPTED {
				DoorTimeout(msg_fromFSM)
			}
		}
	default:

	}
}

func DoorTimeout(msg_fromFSM chan def.MapMessage) {
	switch elevator_state {
	case def.S_DoorOpen:
		currentMap := ordermanager.GetElevMap()
		IO.SetDoorOpenLamp(false)

		motor_direction = ChooseDirection(currentMap)
		IO.SetMotorDirection(motor_direction)

		if motor_direction == IO.MD_Stop {
			elevator_state = def.S_Idle
		} else {
			elevator_state = def.S_Moving
			fmt.Println("func doorTimeout")
		}

		currentMap[def.LOCAL_ID].Dir = motor_direction
		currentMap[def.LOCAL_ID].State = elevator_state
		SendMapMessage(msg_fromFSM, currentMap, nil)
	}
}

func SetButtonLights(currentMap ordermanager.ElevatorMap) {

	for floor := 0; floor < def.NUMFLOORS; floor++ {
		if currentMap[def.LOCAL_ID].Orders[floor][IO.BT_Cab] == ordermanager.ORDER_ACCEPTED {
			IO.SetButtonLamp(IO.BT_Cab, floor, true)
		} else if currentMap[def.LOCAL_ID].Orders[floor][IO.BT_Cab] == ordermanager.NO_ORDER {
			IO.SetButtonLamp(IO.BT_Cab, floor, false)
		}

		if currentMap[def.LOCAL_ID].Orders[floor][IO.BT_HallUp] == ordermanager.ORDER_ACCEPTED {
			IO.SetButtonLamp(IO.BT_HallUp, floor, true)
		} else if currentMap[def.LOCAL_ID].Orders[floor][IO.BT_HallUp] == ordermanager.NO_ORDER {
			IO.SetButtonLamp(IO.BT_HallUp, floor, false)
		}

		if currentMap[def.LOCAL_ID].Orders[floor][IO.BT_HallDown] == ordermanager.ORDER_ACCEPTED {
			IO.SetButtonLamp(IO.BT_HallDown, floor, true)
		} else if currentMap[def.LOCAL_ID].Orders[floor][IO.BT_HallDown] == ordermanager.NO_ORDER {
			IO.SetButtonLamp(IO.BT_HallDown, floor, false)
		}
	}
}

func SendMapMessage(msg_fromFSM chan def.MapMessage, newMap interface{}, newEvent interface{}) {
	sendMsg := def.MakeMapMessage(newMap, newEvent)
	msg_fromFSM <- sendMsg
}
