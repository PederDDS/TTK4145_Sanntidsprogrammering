package fsm

import (
	"fmt"
	"time"

	"../IO"
	"../def"
	"../ordermanager"
)

var elevator_state def.ElevState = def.S_Dead
var motor_direction IO.MotorDirection
var initialized bool = false

func timer(timeout chan<- bool) {
	time.Sleep(5 * time.Second)
	timeout <- true
}

func Initialize(floor_detection <-chan int, fsm_chn chan<- bool, elevator_map_chn chan<- def.MapMessage, direction IO.MotorDirection) {
	currentMap := ordermanager.GetElevMap()

	timeout := make(chan bool, 1)

	elevator_state = def.S_Init
	motor_direction = direction
	IO.SetMotorDirection(motor_direction)

	currentMap[def.LOCAL_ID].State = elevator_state
	currentMap[def.LOCAL_ID].Dir = motor_direction
	currentMap, _ = ordermanager.UpdateElevMap(currentMap)

	go timer(timeout)

	select {
	case floor := <-floor_detection:
		IO.SetFloorIndicator(floor)

		motor_direction = IO.MD_Stop
		elevator_state = def.S_Idle
		IO.SetMotorDirection(motor_direction)

		currentMap[def.LOCAL_ID].Dir = motor_direction
		currentMap[def.LOCAL_ID].State = elevator_state
		currentMap[def.LOCAL_ID].Floor = floor

		currentMap, _ = ordermanager.UpdateElevMap(currentMap)
		time.Sleep(2 * time.Second)
		fsm_chn <- true

	case <-timeout:
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
			currentMap := ordermanager.GetElevMap()
			if currentMap[def.LOCAL_ID].State == def.S_Idle {
				motor_direction = ChooseDirection(currentMap)
				IO.SetMotorDirection(motor_direction)
				currentMap[def.LOCAL_ID].Dir = motor_direction
				if motor_direction != IO.MD_Stop {
					currentMap[def.LOCAL_ID].State = def.S_Moving
					elevator_state = def.S_Moving
				} else if motor_direction == IO.MD_Stop {
					currentMap[def.LOCAL_ID].State = def.S_Idle
					elevator_state = def.S_Idle
				}
				SendMapMessage(msg_fromFSM, currentMap, nil)
			}
		case arrivalFloor := <-drv_floors: //elevator arrives at new floor

			FloorArrival(msg_fromFSM, arrivalFloor, doorTimer)
			idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)

		case msg := <-msg_deadElev: //elevator is dead
			localMap := ordermanager.GetElevMap()
			for elev := 0; elev < def.NUMELEVATORS; elev++ {
				if msg.SendMap.(ordermanager.ElevatorMap)[elev].State == def.S_Dead && localMap[elev].State != def.S_Dead && msg.SendEvent == "Dead elevator" {
					fmt.Println("FSM thinks that the dead elev is:", elev)
					DeadElevator(msg_fromFSM, elev)
					idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)
					if msg.SendMap.(ordermanager.ElevatorMap)[def.LOCAL_ID].State == def.S_Dead {
						elevator_state = def.S_Dead
					}
				} else if msg.SendMap.(ordermanager.ElevatorMap)[elev].State == def.S_Idle && localMap[elev].State == def.S_Dead && msg.SendEvent == "New elevator" {
					localMap := ordermanager.GetElevMap()
					localMap = ordermanager.OverWriteIdle(localMap, elev)
					if msg.SendMap.(ordermanager.ElevatorMap)[def.LOCAL_ID].State != def.S_Dead {
						elevator_state = def.S_Idle
					}

					SendMapMessage(msg_fromFSM, localMap, nil)
				}
			}

		case msg_button := <-drv_buttons: //detects new buttons pushed
			currentMap := ordermanager.GetElevMap()

			if msg_button.Button == IO.BT_Cab {

				currentMap[def.LOCAL_ID].Orders[msg_button.Floor][msg_button.Button] = ordermanager.ORDER

				currentMap, _ = ordermanager.UpdateElevMap(currentMap)
				ButtonPushed(msg_fromFSM, msg_button.Floor, int(msg_button.Button), doorTimer, currentMap)
				idleTimer.Reset(def.IDLE_TIMEOUT_TIME * time.Second)

			} else if msg_button.Button == IO.BT_HallUp || msg_button.Button == IO.BT_HallDown {
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
		if currentMap[def.LOCAL_ID].Floor != def.NUMFLOORS-1 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] == ordermanager.ORDER_ACCEPTED {
				return IO.MD_Stop
			}

			if ordermanager.IsOrderAbove(currentMap) {
				return IO.MD_Up
			}
		}

		if currentMap[def.LOCAL_ID].Floor != 0 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] == ordermanager.ORDER_ACCEPTED {
				return IO.MD_Stop
			}

			if ordermanager.IsOrderBelow(currentMap) {
				return IO.MD_Down
			}
		}

		if currentMap[def.LOCAL_ID].Floor == def.NUMFLOORS-1 {
			return IO.MD_Stop
		}

		return IO.MD_Stop

	case IO.MD_Down:
		if currentMap[def.LOCAL_ID].Floor != 0 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] == ordermanager.ORDER_ACCEPTED {
				return IO.MD_Stop
			}

			if ordermanager.IsOrderBelow(currentMap) {
				return IO.MD_Down
			}
		}

		if currentMap[def.LOCAL_ID].Floor != def.NUMFLOORS-1 {
			if currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] == ordermanager.ORDER_ACCEPTED {
				return IO.MD_Stop
			}

			if ordermanager.IsOrderAbove(currentMap) {
				return IO.MD_Up
			}
		}

		if currentMap[def.LOCAL_ID].Floor == 0 {
			return IO.MD_Stop
		}

		return IO.MD_Stop

	case IO.MD_Stop:
		if ordermanager.IsOrderAbove(currentMap) {
			return IO.MD_Up
		} else if ordermanager.IsOrderBelow(currentMap) {
			return IO.MD_Down
		} else {
			return IO.MD_Stop
		}

	default:
		return IO.MD_Stop
	}
}

func FloorArrival(msg_fromFSM chan def.MapMessage, arrivalFloor int, doorTimer *time.Timer) {
	currentMap := ordermanager.GetElevMap()
	currentMap[def.LOCAL_ID].Floor = arrivalFloor
	IO.SetFloorIndicator(arrivalFloor)

	switch currentMap[def.LOCAL_ID].State {
	case def.S_Idle:

		//if order on floor, delete orders and set door open
		if ordermanager.IsOrderOnFloor(currentMap, arrivalFloor) {
			motor_direction = IO.MD_Stop
			elevator_state = def.S_DoorOpen

			IO.SetMotorDirection(motor_direction)
			IO.SetDoorOpenLamp(true)

			currentMap = ordermanager.DeleteOrdersOnFloor(currentMap, arrivalFloor)
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
		motor_direction = ChooseDirection(currentMap)
		IO.SetMotorDirection(motor_direction)

		if motor_direction == IO.MD_Stop && ordermanager.IsOrderOnFloor(currentMap, arrivalFloor) {

			IO.SetMotorDirection(motor_direction)
			elevator_state = def.S_DoorOpen

			IO.SetDoorOpenLamp(true)
			doorTimer.Reset(def.DOOR_TIMEOUT_TIME * time.Second)
			currentMap := ordermanager.DeleteOrdersOnFloor(currentMap, arrivalFloor)
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
		<-doorTimer.C
		DoorTimeout(msg_fromFSM)
	}
}

func DeadElevator(msg_fromFSM chan def.MapMessage, deadElevId int) {
	currentMap := ordermanager.GetElevMap()
	currentMap = ordermanager.OverWriteDead(currentMap, deadElevId)
	currentMap = ordermanager.RedistributeOrders(currentMap, deadElevId)

	SendMapMessage(msg_fromFSM, currentMap, nil)

}

func ButtonPushed(msg_fromFSM chan def.MapMessage, floor int, button int, doorTimer *time.Timer, currentMap ordermanager.ElevatorMap) {
	SetButtonLights(currentMap)

	switch currentMap[def.LOCAL_ID].State {
	case def.S_Idle:
		if currentMap[def.LOCAL_ID].Floor == floor {
			motor_direction = IO.MD_Stop
			IO.SetMotorDirection(motor_direction)
			currentMap = ordermanager.DeleteOrdersOnFloor(currentMap, floor)

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
			IO.SetMotorDirection(motor_direction)

			if motor_direction != IO.MD_Stop {
				elevator_state = def.S_Moving
			}

			currentMap[def.LOCAL_ID].Dir = motor_direction
			currentMap[def.LOCAL_ID].State = elevator_state
			SendMapMessage(msg_fromFSM, currentMap, nil)
		}

	case def.S_DoorOpen:
		if currentMap[def.LOCAL_ID].Floor == floor {
			currentMap = ordermanager.DeleteOrdersOnFloor(currentMap, floor)
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
