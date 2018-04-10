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
var initialized bool = false

func timer(timeout chan<- bool){
  fmt.Println("Timer started")
  time.Sleep(5*time.Second)
  timeout <- true
  fmt.Println("Timeout sent")
}

func Initialize(floor_detection <-chan int, fsm_chn chan<- bool, elevator_map_chn chan<- def.MapMessage, direction IO.MotorDirection){
  currentMap := ordermanager.GetElevMap()

  timeout := make(chan bool, 1)
  elevator_state = def.S_Init
  currentMap[def.LOCAL_ID].State = elevator_state
  motor_direction = direction
  currentMap[def.LOCAL_ID].Dir = motor_direction
  sendMessage := def.MakeMapMessage(currentMap, nil)

  newMap, _ := ordermanager.UpdateElevMap(sendMessage.SendMap.(ordermanager.ElevatorMap))
  IO.SetMotorDirection(motor_direction)
  go timer(timeout)
  newMap[1].Dir = 0 // fordi newMap is declared and not used...

  select{
  case <- floor_detection:
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
    sendMessage = def.MakeMapMessage(currentMap, nil)

    newMap, _ = ordermanager.UpdateElevMap(sendMessage.SendMap.(ordermanager.ElevatorMap))
    time.Sleep(2*time.Second)
    fsm_chn <- true
  case <- timeout:
  fmt.Println("Timed out")
  Initialize(floor_detection, fsm_chn, elevator_map_chn, - motor_direction)
  }
}

func Dust(msg_fromFSM chan def.MapMessage){
  fmt.Println("func: Dust")
  currentMap := ordermanager.GetElevMap()
  currentMap[def.LOCAL_ID].State = def.S_Dead
  message := def.MakeMapMessage(currentMap, nil)
  msg_fromFSM <- message
}


func FSM(drv_buttons <-chan IO.ButtonEvent, drv_floors <-chan int, fsm_chn chan bool, elevator_map_chn chan def.MapMessage, direction IO.MotorDirection, msg_buttonEvent chan def.MapMessage, msg_fromHWFloor chan def.MapMessage, msg_fromHWButton chan def.MapMessage, msg_fromFSM chan def.MapMessage, msg_deadElev chan def.MapMessage) {

    if initialized == false {
        Initialize(drv_floors, fsm_chn, elevator_map_chn, IO.MD_Up)
        <- fsm_chn
        initialized = true
        fmt.Println("Elevator is initialized!")
    }


    doorTimer := time.NewTimer(def.DOOR_TIMEOUT_TIME*time.Second)
    doorTimer.Stop()
    idleTimer := time.NewTimer(def.IDLE_TIMEOUT_TIME*time.Second)

    for {
      select {
      case msg := <- msg_fromHWFloor:
          switch msg.SendEvent.(def.NewEvent).EventType {
          case def.FLOOR_ARRIVAL:
            FloorArrival(msg_fromFSM, msg.SendEvent.(def.NewEvent).Type.(int), doorTimer)
            idleTimer.Reset(def.IDLE_TIMEOUT_TIME*time.Second)
          }

      case msg := <- msg_buttonEvent:
        switch msg.SendEvent.(def.NewEvent).EventType {
        case def.BUTTON_PUSHED:
          button := msg.SendEvent.(def.NewEvent).Type.([]int)
          ButtonPushed(msg_fromFSM, button[0], button[1], doorTimer)
          idleTimer.Reset(def.IDLE_TIMEOUT_TIME*time.Second)
        }

      case msg := <- msg_deadElev:
        switch msg.SendEvent.(def.NewEvent).EventType {
        case def.ELEVATOR_DEAD:
          DeadElevator(msg_fromFSM, msg.SendEvent.(def.NewEvent).Type.(int))
          idleTimer.Reset(def.IDLE_TIMEOUT_TIME*time.Second)
        }

      default:

    }
  }
}


func ChooseDirection(currentMap ordermanager.ElevatorMap) IO.MotorDirection {
  switch motor_direction {
  case IO.MD_Up:
    return IO.MD_Up
  }
  return IO.MD_Down
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
    currentMap[elev].Buttons[currentFloor][IO.BT_HallUp] = 0
    currentMap[elev].Buttons[currentFloor][IO.BT_HallDown] = 0
  }
  currentMap[def.LOCAL_ID].Buttons[currentFloor][IO.BT_Cab] = 0

  for button := 0; button < def.NUMBUTTON_TYPES; button++ {
    currentMap[def.LOCAL_ID].Orders[currentFloor][button] = 0
  }
return currentMap
}


func IsOrderOnFloor(currentMap ordermanager.ElevatorMap, currentFloor int) bool {
  if currentMap[def.LOCAL_ID].Buttons[currentFloor][IO.BT_Cab] == 1 {
    return true
  }

  for elev := 0; elev < def.NUMELEVATORS; elev++ {
    if currentMap[elev].Buttons[currentFloor][IO.BT_HallUp] == 1 || currentMap[elev].Buttons[currentFloor][IO.BT_HallDown] == 1 {
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


func FloorArrival(msg_fromFSM chan def.MapMessage, arrivalFloor int, doorTimer *time.Timer) {
  currentMap := ordermanager.GetElevMap()
  currentMap[def.LOCAL_ID].Floor = arrivalFloor

  switch elevator_state {
  case def.S_Idle:
    sendMsg := def.MakeMapMessage(currentMap, nil)
    msg_fromFSM <- sendMsg

  case def.S_Moving:
    if PossibleStopOnFloor(currentMap) {
      IO.SetMotorDirection(IO.MD_Stop)
      IO.SetDoorOpenLamp(true)
      IO.SetFloorIndicator(arrivalFloor)

      motor_direction = IO.MD_Stop
      elevator_state = def.S_DoorOpen
      currentMap[def.LOCAL_ID].State = elevator_state

      currentMap := DeleteOrdersOnFloor(currentMap, arrivalFloor)

      doorTimer.Reset(def.DOOR_TIMEOUT_TIME*time.Second)

      sendMsg := def.MakeMapMessage(currentMap, nil)
      msg_fromFSM <- sendMsg
    }
  }
}


func DeadElevator(msg_fromFSM chan def.MapMessage, deadElevID int) {

}


func ButtonPushed(msg_fromFSM chan def.MapMessage, floor int, button int, doorTimer *time.Timer) {

}


func DoorTimeout(msg_fromFSM chan def.MapMessage){
  switch elevator_state {
  case def.S_DoorOpen:
    currentMap := ordermanager.GetElevMap()
    IO.SetDoorOpenLamp(false)

    motor_direction = ChooseDirection(currentMap)
    IO.SetMotorDirection(motor_direction)
    currentMap[def.LOCAL_ID].Dir = motor_direction

    if motor_direction == IO.MD_Stop {
      elevator_state = def.S_Idle
    } else {
      elevator_state = def.S_Moving
    }

    SendMapMessage(msg_fromFSM, currentMap, nil)
  }
}


func PossibleStopOnFloor(currentMap ordermanager.ElevatorMap) bool {
  floor := currentMap[def.LOCAL_ID].Floor
  switch motor_direction {
  case IO.MD_Up:
    if currentMap[def.LOCAL_ID].Buttons[floor][IO.BT_HallUp] == 1 {
      return true
    }
  }
  return false
}


func SendMapMessage(msg_fromFSM chan def.MapMessage, newMap interface{}, newEvent interface{} ) {
  sendMsg := def.MakeMapMessage(newMap, nil)
  msg_fromFSM <- sendMsg
}
