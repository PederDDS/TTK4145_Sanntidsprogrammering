package fsm

import (
  "fmt"
  "time"
  "../def"
  "../ordermanager"
  "../IO"
)


var elevator_state def.ElevState = def.S_Dead
var motor_direction IO.MotorDirection
var init bool = false

func timer(timeout chan<- bool){
  fmt.Println("Timer started")
  time.Sleep(5*time.Second)
  timeout <- true
  fmt.Println("Timeout sent")
}

func Initialize(floor_detection <-chan int, to_main chan<- bool, elevator_map_chn chan<- def.MapMessage, direction IO.MotorDirection){
  timeout := make(chan bool, 1)
  elevator_state = def.S_Init
  motor_direction = direction
  sendMessage := def.MakeMapMessage(nil, motor_direction)
  IO.SetMotorDirection(motor_direction)
  go timer(timeout)
  select{
  case <- floor_detection:
    fmt.Println("Floor detected")
    motor_direction = IO.MD_Stop
    IO.SetMotorDirection(motor_direction)
    elevator_state = def.S_Idle
    time.Sleep(2*time.Second)
    to_main <- true
  case <- timeout:
  fmt.Println("Timed out")
  Initialize(floor_detection, to_main, - motor_direction)
  }
}

func PrintState(){
  for{
    time.Sleep(2*time.Second)
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

drv_buttons := make(chan IO.ButtonEvent)
drv_floors  := make(chan int)
drv_obstr   := make(chan bool)
drv_stop    := make(chan bool)
bcast_chn		:= make(chan IO.ButtonEvent)
recieve_chn	:= make(chan string)
fsm_chn			:= make(chan bool, 1)

func FSM(drv_buttons <-chan IO.ButtonEvent, drv_floors <-chan int, fsm_chn chan bool, elevator_map_chn chan def.MapMessage) {
    if init == false {
        Initialize(drv_floors, fsm_chn, elevator_map_chn, IO.MD_Up)
        init = true
    }

    switch elevator_state{
    case def.S_Dead:
        // Initialize again, but in opposite direction? -> S_Init
        Initialize(drv_floors, fsm_chn, - motor_direction)
    case def.S_Init:
        // Check for orders         -> S_Moving
        // If order on floor        -> S_DoorOpen
        // If no orders             -> S_Idle
        // If unable to initialize  -> S_Init
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


//func ChooseDirection() int {
// return 0
//}


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
    currentMap[elev].Buttons[currentFloor][def.BT_HallUp] = 0
    currentMap[elev].Buttons[currentFloor][def.BT_HallDown] = 0
  }
  currentMap[def.LOCAL_ID].Buttons[currentFloor][def.BT_Cab] = 0

  for button := 0; button < def.NUMBUTTON_TYPES; button++ {
    currentMap[def.LOCAL_ID].Orders[currentFloor][button] = 0
  }
return currentMap
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
    if currentMap[def.LOCAL_ID].Orders[currentFloor][button] == 1 {
      return true
    }
  }
  return false
}
