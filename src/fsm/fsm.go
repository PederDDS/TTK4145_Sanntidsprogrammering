package fsm

import (
  "fmt"
  "time"
  "../def"
  //"../ordermanager"
  "../IO"
)

var elevator_state def.ElevState = def.S_Dead
var motor_direction IO.MotorDirection

func timer(timeout chan<- bool){
  fmt.Println("Timer started")
  time.Sleep(5*time.Second)
  timeout <- true
  fmt.Println("Timeout sent")
}

func Initialize(floor_detection <-chan int, direction IO.MotorDirection){
  timeout := make(chan bool)
  elevator_state = def.S_Init
  motor_direction = direction
  IO.SetMotorDirection(motor_direction)
  go timer(timeout)
  select{
  case <- floor_detection:
    fmt.Println("Floor detected")
    motor_direction = IO.MD_Stop
    IO.SetMotorDirection(motor_direction)
    elevator_state = def.S_Idle
  case <- timeout:
    fmt.Println("Timed out")
    Initialize(floor_detection, - motor_direction)
  }
}

func PrintState(){
  for{
    time.Sleep(time.Second)
    switch elevator_state{
    case def.S_Dead:
      fmt.Println("State: Dead")
    case def.S_Init:
      fmt.Println("State: Initializing")
    case def.S_Idle:
      fmt.Println("State: Idle")
    case def.S_Moving:
      fmt.Println("State: Moving")
    case def.S_DoorOpen:
      fmt.Println("State: Door Open")
    }
  }
}


func FSM(/*Lots of channels*/) {
  switch elevator_state{
  case def.S_Dead:
    // Initialize again, but in opposite direction? -> S_Init
  case def.S_Init:
    // Check for orders         -> S_Moving
    // If order on floor        -> S_DoorOpen
    // If no orders             -> S_Idle
    // If unable to initialize  -> S_Dead
  case def.S_Idle:
    // Check for orders   -> S_Moving
    // If order on floor  -> S_DoorOpen
  case def.S_Moving:
    // Check for orders on passing floors                           -> S_DoorOpen
    // If unable to reach floor after a reasonable amount of time   -> S_Dead
  case def.S_DoorOpen:
    // Check for orders -> S_Moving
    // If no orders     -> S_Idle
  }
}
/*

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
