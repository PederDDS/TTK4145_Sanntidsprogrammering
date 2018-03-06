package main

import (
	"../def"
	"../IO"
	//"net"
	//"../network/bcast"
	//"../network/localip"
	// "flag"
	"fmt"
	//"os"
	//"time"
	// "time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {

	    IO.Init("localhost:15657", def.NUMFLOORS)

	    var d IO.MotorDirection = IO.MD_Up
	    //IO.SetMotorDirection(d)

	    drv_buttons := make(chan IO.ButtonEvent)
	    drv_floors  := make(chan int)
	    drv_obstr   := make(chan bool)
	    drv_stop    := make(chan bool)

	    go IO.PollButtons(drv_buttons)
	    go IO.PollFloorSensor(drv_floors)
	    go IO.PollObstructionSwitch(drv_obstr)
	    go IO.PollStopButton(drv_stop)


	    for {
	        select {
	        case a := <- drv_buttons:
	            fmt.Printf("%+v\n", a)
	            IO.SetButtonLamp(a.Button, a.Floor, true)

	        case a := <- drv_floors:
	            fmt.Printf("%+v\n", a)
	            if a == def.NUMFLOORS-1 {
	                d = IO.MD_Down
	            } else if a == 0 {
	                d = IO.MD_Up
	            }
	            IO.SetMotorDirection(d)


	        case a := <- drv_obstr:
	            fmt.Printf("%+v\n", a)
	            if a {
	                IO.SetMotorDirection(IO.MD_Stop)
	            } else {
	                IO.SetMotorDirection(d)
	            }

	        case a := <- drv_stop:
	            fmt.Printf("%+v\n", a)
	            for f := 0; f < def.NUMFLOORS; f++ {
	                for b := IO.ButtonType(0); b < 3; b++ {
	                    IO.SetButtonLamp(b, f, false)
	                }
	            }
	        }
	    }
	}
