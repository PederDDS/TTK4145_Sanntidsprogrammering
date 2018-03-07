package fsm

import (
  "fmt"
  //"time"
  "../def"
  //"../ordermanager"
  "../IO"
)

var elevator_state def.ElevState
var motor_direction IO.MotorDirection

func Initialize(floor_detection <-chan int){
    fmt.Println("Initializing")
    elevator_state = def.S_Init
    motor_direction = IO.MD_Up
    IO.SetMotorDirection(motor_direction)
    select{
    case <- floor_detection:
      fmt.Println("Noe skjedde")
      motor_direction = IO.MD_Stop
      IO.SetMotorDirection(motor_direction)
      elevator_state = def.S_Idle
    }
  }

/*
func FSM() {

}


func ChooseDirection() int {

}


func OrderAbove(currentMap ordermanager.ElevatorMap) bool {
  for floor := currentMap[def.LOCAL_ID].Floor + 1; floor < def.NUMFLOORS; floor++ {
    if IsOrderOnFloor(currentMap, floor) {
      return true
    }
  }
  return false
}


func OrderBelow(currentMap ordermanager.ElevatorMap) bool {
  for floor := 0; floor < currentMap[def.LOCAL_ID].Floor; floor++ {
    if IsOrderOnFloor(currentMap, floor) {
      return true
    }
  }
  return false
}


func DeleteOrdersOnFloor(currentMap ordermanager.ElevatorMap, currentFloor int) ordermanager.ElevatorMap {
  for elev := 0; elev < def.NUMELEVATORS; elev++ {
    currentMap[elev].Buttons[floor][def.BT_HallUp] = 0
    currentMap[elev].Buttons[floor][def.BT_HallDown] = 0
  }
  currentMap[def.LOCAL_ID].Buttons[floor][def.BT_Cab] = 0

  for button := 0; button < def.NUMBUTTON_TYPES; button++ {
    currentMap[def.LOCAL_ID].Orders[floor][button] = 0
  }
}


func IsOrderOnFloor(currentMap ordermanager.ElevatorMap, currentFloor int) bool {
  if currentMap[def.LOCAL_ID].Buttons[currentFloor][def.BT_Cab] == 1 {
    return true
  }

  for elev := 0; elev < def.NUMELEVATORS; elev++ {
    if currentMap[elev].Buttons[currentFloor][def.BT_HallUp] == 1 || currentMap[elev].Buttons[currentFloor][def.BT_HallDown] == 1 {
      return true
    }
  }

  for button := 0; button < def.NUMBUTTON_TYPES; button++ {
    if currentMap[def.LOCAL_ID].Orders[floor][button] == 1 {
      return true
    }
  }
  return false
}
*/
