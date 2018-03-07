package fsm

import (
  "../def"
  "../IO"
  "fmt"
)

var elevator_state def.ElevState
var motor_direction IO.MotorDirection

func Initialize(floor_detection <-chan int,to_main chan<- bool){
    fmt.Println("Initializing")
    elevator_state = def.S_Init
    motor_direction = IO.MD_Up
    IO.SetMotorDirection(motor_direction)
    floor := <- floor_detection
    loop:
      for{
        fmt.Println("looping")
        if floor != <-floor_detection{
          break loop
        }
      }
    to_main <- true
    motor_direction = IO.MD_Stop
    IO.SetMotorDirection(motor_direction)
    elevator_state = def.S_Idle
}
