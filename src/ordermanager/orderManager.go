package ordermanager

import (
	"fmt"
	"math"
	"sync"

	"../IO"
	"../def"
)

var mapMtx = &sync.Mutex{}
var localElevMap *ElevatorMap

//structs and constants
type ElevatorMap [def.NUMELEVATORS]Elev

type Elev struct {
	ElevID  int
	Active	bool
	Dir     IO.MotorDirection
	Floor   int
	State   def.ElevState
	Buttons [def.NUMFLOORS][def.NUMBUTTON_TYPES]int
	Orders  [def.NUMFLOORS][def.NUMBUTTON_TYPES]int
}

type Request int

const (
	NO_ORDER         = 0
	ORDER            = 1
	ORDER_ACCEPTED   = 2
	ORDER_IMPOSSIBLE = -1

	LAMP_OFF = 0
	LAMP_ON  = 1
)

type ElevRequests []Request

//functions
func InitElevMap(backup bool) {
	mapMtx.Lock()
	localElevMap = new(ElevatorMap)

	if backup {
		*localElevMap = GetBackup()
	} else {
		localElevMap = MakeEmptyElevMap()
	}

	MakeBackup(*localElevMap)
	mapMtx.Unlock()
}

func PrintElevMap() {
	// Husk å endre 1 tilbake til def.NUMELEVATORS!!
	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		fmt.Println("-----------------------------------------------")
		fmt.Println("Elevator number:", localElevMap[elev].ElevID)
		fmt.Println("Active elevator:", localElevMap[elev].Active)
		switch localElevMap[elev].Dir {
		case -1:
			fmt.Println("Motor direction: Down")
		case 0:
			fmt.Println("Motor direction: Stop")
		case 1:
			fmt.Println("Motor direction: Up")
		}

		switch localElevMap[elev].Floor {
		case 0:
			fmt.Println("Floor: 1st")
		case 1:
			fmt.Println("Floor: 2nd")
		case 2:
			fmt.Println("Floor: 3rd")
		default:
			fmt.Print("Floor: ", localElevMap[elev].Floor+1, "th\n")
		}

		switch localElevMap[elev].State {
		case 0:
			fmt.Println("State: Dead")
		case 1:
			fmt.Println("State: Initializing")
		case 2:
			fmt.Println("State: Idle")
		case 3:
			fmt.Println("State: Moving")
		case 4:
			fmt.Println("State: Door open")
		}
		fmt.Println("Buttons:    U D C")
		fmt.Println("4th floor:", localElevMap[elev].Buttons[3])
		fmt.Println("3rd floor:", localElevMap[elev].Buttons[2])
		fmt.Println("2nd floor:", localElevMap[elev].Buttons[1])
		fmt.Println("1st floor:", localElevMap[elev].Buttons[0])
		fmt.Println("Orders:     U D C")
		fmt.Println("4th floor:", localElevMap[elev].Orders[3])
		fmt.Println("3rd floor:", localElevMap[elev].Orders[2])
		fmt.Println("2nd floor:", localElevMap[elev].Orders[1])
		fmt.Println("1st floor:", localElevMap[elev].Orders[0])
	}
}

func UpdateElevMap(newMap ElevatorMap) (ElevatorMap, bool) {
	currentMap := GetElevMap()
	allChangesMade := false

	//update buttons and orders
	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		//update directions
		if newMap[elev].Active {
			if newMap[elev].Dir != currentMap[elev].Dir {
				currentMap[elev].Dir = newMap[elev].Dir
				allChangesMade = true
			}

			//update floors
			if newMap[elev].Floor != currentMap[elev].Floor {
				currentMap[elev].Floor = newMap[elev].Floor
				allChangesMade = true
			}

			for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES; button++ {
					if newMap[elev].Buttons[floor][button] != currentMap[elev].Buttons[floor][button] {
						currentMap[elev].Buttons[floor][button] = newMap[elev].Buttons[floor][button]
						allChangesMade = true
					}
					if newMap[elev].Orders[floor][button] != currentMap[elev].Orders[floor][button] {
						currentMap[elev].Orders[floor][button] = newMap[elev].Orders[floor][button]
						allChangesMade = true
					}

				}
			}
		}
		//update states
		if newMap[elev].State != currentMap[elev].State && (newMap[elev].State != def.S_Dead || elev == def.LOCAL_ID) {
			currentMap[elev].State = newMap[elev].State
			allChangesMade = true
		}
	}
	//handle cab orders
	for floor := 0; floor < def.NUMFLOORS; floor++ {
		for button := 0; button < def.NUMBUTTON_TYPES; button++ {
			if button == IO.BT_Cab {
				if newMap[def.LOCAL_ID].Orders[floor][button] == ORDER && currentMap[def.LOCAL_ID].State != def.S_Dead {
					currentMap[def.LOCAL_ID].Orders[floor][button] = ORDER_ACCEPTED
					currentMap[def.LOCAL_ID].Buttons[floor][button] = LAMP_ON
					allChangesMade = true
				} else if newMap[def.LOCAL_ID].Orders[floor][button] == NO_ORDER && currentMap[def.LOCAL_ID].State != def.S_Dead {
					currentMap[def.LOCAL_ID].Orders[floor][button] = NO_ORDER
					currentMap[def.LOCAL_ID].Buttons[floor][button] = LAMP_OFF
					allChangesMade = true
				}
			}
		}
	}

	//check if all elevators are alive
	var tempElevAlive = 0
	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		if currentMap[elev].State != def.S_Dead && currentMap[elev].State != def.S_Init {
			tempElevAlive++
		}
	}

	//handle hall orders
	if tempElevAlive == def.NUMELEVATORS {
		for elev := 0; elev < def.NUMELEVATORS; elev++ {
			for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES-1; button++ {
					if currentMap[elev].Orders[floor][button] == ORDER {
						currentMap = SetToOrder(currentMap, ORDER_ACCEPTED, IO.ButtonType(button))
						allChangesMade = true
					}
				}
			}
		}
	}
	if tempElevAlive == def.NUMELEVATORS {
		currentMap = DistributeOrders(currentMap)
	}

	MakeBackup(currentMap)
	SetElevMap(currentMap)

	PrintElevMap()
	return currentMap, allChangesMade
}

func SetToOrder(currentMap ElevatorMap, order int, button IO.ButtonType) ElevatorMap {
	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			if currentMap[elev].State != def.S_Dead {
				currentMap[elev].Orders[floor][button] = order
			}
		}
	}
	return currentMap
}

func DistributeOrders(currentMap ElevatorMap) ElevatorMap {
	for floor := 0; floor < def.NUMFLOORS; floor++ {
		for button := 0; button < def.NUMBUTTON_TYPES-1; button++ {
			if IsClosestElevator(currentMap, floor) {
				currentMap = SetToOrder(currentMap, NO_ORDER, IO.ButtonType(button))
				currentMap[def.LOCAL_ID].Orders[floor][button] = ORDER_ACCEPTED
			}
		}
	}
	return currentMap
}

func RedistributeOrders(currentMap ElevatorMap, deadElevId int) ElevatorMap {
	for floor := 0; floor < def.NUMFLOORS; floor++ {
		for button := 0; button < def.NUMBUTTON_TYPES; button++ {
			if currentMap[deadElevId].Orders[floor][button] == ORDER_ACCEPTED {
				DistributeOrders(currentMap)
			}
		}

	}
	return currentMap
}

func GetNewEvent(newMap ElevatorMap) (ElevatorMap, [][]int) {
	currentMap := GetElevMap()
	var buttonChanges [][]int

	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		if currentMap[def.LOCAL_ID].State != def.S_Dead && currentMap[def.LOCAL_ID].State != def.S_Init && newMap[def.LOCAL_ID].State != def.S_Dead && newMap[def.LOCAL_ID].State != def.S_Init {

			for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES; button++ {

					if newMap[elev].Buttons[floor][button] == 1 && currentMap[elev].Buttons[floor][button] != 1 {
						if button != IO.BT_Cab {
							currentMap[elev].Buttons[floor][button] = newMap[elev].Buttons[floor][button]
							buttonChanges = append(buttonChanges, []int{floor, button})
						} else {
							currentMap[elev].Buttons[floor][button] = newMap[elev].Buttons[floor][button]
						}
					}
				}
			}
		}
	}

	MakeBackup(currentMap)
	SetElevMap(currentMap)

	return currentMap, buttonChanges
}

func SetElevMap(newMap ElevatorMap) {
	mapMtx.Lock()
	*localElevMap = newMap
	mapMtx.Unlock()
}

func GetElevMap() ElevatorMap {
	mapMtx.Lock()
	elevMap := *localElevMap
	mapMtx.Unlock()
	return elevMap
}

func MakeEmptyElevMap() *ElevatorMap {
	emptyMap := new(ElevatorMap)

	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		emptyMap[elev].ElevID = elev
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			for button := 0; button < def.NUMBUTTON_TYPES; button++ {
				emptyMap[elev].Buttons[floor][button] = 0
				emptyMap[elev].Orders[floor][button] = NO_ORDER
			}
		}
		emptyMap[def.LOCAL_ID].State = def.S_Dead
		emptyMap[def.LOCAL_ID].Active = true
		emptyMap[def.LOCAL_ID].Dir = IO.MD_Stop
		emptyMap[def.LOCAL_ID].Floor = -1
	}
	PrintElevMap()
	return emptyMap
}

func IsClosestElevator(currentMap ElevatorMap, floor int) bool {
	result := true
	myDistance := int(math.Abs(float64(currentMap[def.LOCAL_ID].Floor - floor)))

	if currentMap[def.LOCAL_ID].Floor < floor {
		if currentMap[def.LOCAL_ID].State == def.S_Dead {
			result = false
		}

		for elev := 0; elev < def.NUMELEVATORS; elev++ {

			if elev != def.LOCAL_ID && (currentMap[elev].State != def.S_Dead || currentMap[elev].State != def.S_Init) {

				elevDistance := int(math.Abs(float64(currentMap[elev].Floor - floor)))

				if elevDistance < myDistance {

					if currentMap[elev].Floor < floor && (currentMap[elev].Dir == IO.MD_Up || currentMap[elev].Dir == IO.MD_Stop) {
						result = false
					} else if currentMap[elev].Floor > floor && (currentMap[elev].Dir == IO.MD_Down || currentMap[elev].Dir == IO.MD_Stop) {
						result = false
					} else if currentMap[elev].Floor == floor && currentMap[elev].Dir == IO.MD_Stop {
						result = false
					}

				} else if elevDistance == myDistance && (currentMap[elev].Dir == IO.MD_Up || currentMap[elev].Dir == IO.MD_Stop) {
					if elev < def.LOCAL_ID {
						result = false
					}
				}
			}
		}
	} else if currentMap[def.LOCAL_ID].Floor > floor {
		if currentMap[def.LOCAL_ID].State == def.S_Dead {
			result = false
		}

		for elev := 0; elev < def.NUMELEVATORS; elev++ {

			if elev != def.LOCAL_ID && (currentMap[elev].State != def.S_Dead || currentMap[elev].State != def.S_Init) {

				elevDistance := int(math.Abs(float64(currentMap[elev].Floor - floor)))

				if elevDistance < myDistance {

					if currentMap[elev].Floor < floor && (currentMap[elev].Dir == IO.MD_Up || currentMap[elev].Dir == IO.MD_Stop) {
						result = false
					} else if currentMap[elev].Floor > floor && (currentMap[elev].Dir == IO.MD_Down || currentMap[elev].Dir == IO.MD_Stop) {
						result = false
					} else if currentMap[elev].Floor == floor && currentMap[elev].Dir == IO.MD_Stop {
						result = false

					}

				} else if elevDistance == myDistance && (currentMap[elev].Dir == IO.MD_Down || currentMap[elev].Dir == IO.MD_Stop) {
					if currentMap[elev].ElevID < currentMap[def.LOCAL_ID].ElevID {
						result = false
					}
				}

			}
		}
	}

	return result
}
