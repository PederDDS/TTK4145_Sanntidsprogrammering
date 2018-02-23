package IO

import (
	"def"
	"orderManager"
	"fmt"
	"net"
	"log"
	"time"
	"sync"
)


const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn





//init function
func initHw (addr strin, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}



//set functions
func setElevLights() { //uses setOrderLight() and setFloorIndicatorLight()

}


func setDoorLight(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}


func setFloorIndicatorLight(floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{3, byte(floor), 0, 0})
}


func setButtonLight(button ButtonType, floor int, value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
}


func setMotorDir(dir MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}



//get functions
func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}


func getButton(button ButtonType, floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}



//poll functions
func pollFloors(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		floor := getFloor()
		if floor != prev && floor != -1 {
			receiver <- floor
		}
		prev = floor
	}
}


func pollElevButtons(receiver chan<- ButtonEvent)  {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for floor := 0; floor < _numFloors; floor++ {
			for button := ButtonType(0); button < 3; button++ {
				v := getButton(button, floor)
				if v != prev[floor][button] && v != false {
					receiver <- ButtonEvent{floor, ButtonType(button)}
				}
				prev[floor][button] = v
			}
		}
	}
}




//bits, bytes and bools
func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}


func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
