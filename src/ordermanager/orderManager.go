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

type ElevatorMap [def.NUMELEVATORS]Elev

type Elev struct {
	ElevID  int
	Active	bool
	Dir     IO.MotorDirection
	Floor   int
	State   def.ElevState
	Orders  [def.NUMFLOORS][def.NUMBUTTON_TYPES]int
}


const (
	NO_ORDER         = 0
	ORDER            = 1
	ORDER_ACCEPTED   = 2
	ORDER_IMPOSSIBLE = -1

	LAMP_OFF = 0
	LAMP_ON  = 1
)

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

	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		//only make changes for the elev that sent newMap
		if newMap[elev].Active {
			if newMap[elev].Dir != currentMap[elev].Dir {
				currentMap[elev].Dir = newMap[elev].Dir
				allChangesMade = true
			}

			if newMap[elev].Floor != currentMap[elev].Floor {
				currentMap[elev].Floor = newMap[elev].Floor
				allChangesMade = true
			}

			for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES-1; button++ {
					if currentMap[elev].Orders[floor][button] != ORDER_IMPOSSIBLE {
						 if newMap[elev].Orders[floor][button] == ORDER && currentMap[elev].Orders[floor][button] != ORDER_ACCEPTED {
							 	currentMap[elev].Orders[floor][button] = ORDER
								allChangesMade = true
							}
							if newMap[elev].Orders[floor][button] == ORDER_ACCEPTED && currentMap[elev].Orders[floor][button] != NO_ORDER {
									currentMap[elev].Orders[floor][button] = ORDER_ACCEPTED
									allChangesMade = true
							}
							if newMap[elev].Orders[floor][button] == NO_ORDER && currentMap[elev].Orders[floor][button] == ORDER {
									currentMap[elev].Orders[floor][button] = NO_ORDER
									allChangesMade = true
							}
						}
					}
				}
			}

			if newMap[elev].State != currentMap[elev].State && newMap[elev].State != def.S_Dead && currentMap[elev].State != def.S_Dead {
				currentMap[elev].State = newMap[elev].State
				allChangesMade = true
			}
		}

		if newMap[def.LOCAL_ID].State != currentMap[def.LOCAL_ID].State {
			currentMap[def.LOCAL_ID].State = newMap[def.LOCAL_ID].State
			allChangesMade = true
		}

		//handle cab orders
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			if newMap[def.LOCAL_ID].Orders[floor][IO.BT_Cab] == ORDER && currentMap[def.LOCAL_ID].State != def.S_Dead {
				currentMap[def.LOCAL_ID].Orders[floor][IO.BT_Cab] = ORDER_ACCEPTED
				allChangesMade = true
			}
		}

		var distributeOrders = false

		for elev := 0; elev < def.NUMELEVATORS; elev++ {
			for floor := 0; floor < def.NUMFLOORS; floor ++ {
				for button := 0; button < def.NUMBUTTON_TYPES - 1; button++ {
					if (elev != def.LOCAL_ID && currentMap[elev].Orders[floor][button] == ORDER) && currentMap[def.LOCAL_ID].Orders[floor][button] == NO_ORDER {
						currentMap[def.LOCAL_ID].Orders[floor][button] = ORDER
					}
					ordercounter := 0
					for elevator := 0; elevator < def.NUMELEVATORS; elevator++ {
						if currentMap[def.LOCAL_ID].Orders[floor][button] != ORDER_IMPOSSIBLE && (currentMap[elevator].Orders[floor][button] == ORDER || currentMap[elevator].Orders[floor][button] == ORDER_IMPOSSIBLE){
								ordercounter++
							}
						}
						if ordercounter == def.NUMELEVATORS {
							distributeOrders = true
						}
					}
				}
			}

		if distributeOrders {
			currentMap = DistributeOrders(currentMap)
			for elev := 0; elev < def.NUMELEVATORS; elev++ {
				for floor := 0; floor < def.NUMFLOORS; floor++ {
					for button := 0; button < def.NUMBUTTON_TYPES-1; button++ {
						if currentMap[elev].Orders[floor][button] == ORDER {
							currentMap[elev].Orders[floor][button] = NO_ORDER
						}
					}
				}
			}
		}

	MakeBackup(currentMap)
	SetElevMap(currentMap)
	PrintElevMap()
	return currentMap, allChangesMade
}

func OverWriteDead(newMap ElevatorMap, deadElevId int) ElevatorMap {
	newMap[deadElevId].State = def.S_Dead

	for floor := 0; floor < def.NUMFLOORS; floor++ {
		newMap[deadElevId].Orders[floor][IO.BT_HallUp] = ORDER_IMPOSSIBLE
		newMap[deadElevId].Orders[floor][IO.BT_HallDown] = ORDER_IMPOSSIBLE
	}

	SetElevMap(newMap)
	PrintElevMap()
	return newMap
}

func OverWriteIdle(newMap ElevatorMap, idleElevId int) ElevatorMap {
	newMap[idleElevId].State = def.S_Idle

	for floor := 0; floor < def.NUMFLOORS; floor++ {
		newMap[idleElevId].Orders[floor][IO.BT_HallUp] = NO_ORDER
		newMap[idleElevId].Orders[floor][IO.BT_HallDown] = NO_ORDER
	}

	SetElevMap(newMap)
	PrintElevMap()
	return newMap
}

func SetToOrder(currentMap ElevatorMap, order int, floor int, button IO.ButtonType) ElevatorMap {
	for elev := 0; elev < def.NUMELEVATORS; elev++ {
			if currentMap[elev].State != def.S_Dead {
				currentMap[elev].Orders[floor][button] = order
		}
	}
	return currentMap
}

func IsOrderAbove(currentMap ElevatorMap) bool {
	for floor := currentMap[def.LOCAL_ID].Floor + 1; floor < def.NUMFLOORS; floor++ {
		if IsOrderOnFloor(currentMap, floor) {
			return true
		}
	}
	return false
}

func IsOrderBelow(currentMap ElevatorMap) bool {
	for floor := 0; floor < currentMap[def.LOCAL_ID].Floor; floor++ {
		if IsOrderOnFloor(currentMap, floor) {
			return true
		}
	}
	return false
}

func IsOrderOnFloor(currentMap ElevatorMap, currentFloor int) bool {
	for button := 0; button < def.NUMBUTTON_TYPES; button++ {
		if currentMap[def.LOCAL_ID].Orders[currentFloor][button] == ORDER_ACCEPTED {
			return true
		}
	}
	return false
}

func DistributeOrders(currentMap ElevatorMap) ElevatorMap {
	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			for button := 0; button < def.NUMBUTTON_TYPES-1; button++ {
				if currentMap[elev].Orders[floor][button] == ORDER {
					if IsClosestElevator(currentMap, floor) {
						currentMap = SetToOrder(currentMap, NO_ORDER, floor, IO.ButtonType(button))
						currentMap[def.LOCAL_ID].Orders[floor][button] = ORDER_ACCEPTED
					}
				}
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


func DeleteOrdersOnFloor(currentMap ElevatorMap, currentFloor int) ElevatorMap {
	if currentMap[def.LOCAL_ID].Dir == IO.MD_Up || currentMap[def.LOCAL_ID].Dir == IO.MD_Stop {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] = NO_ORDER
	}
	if currentMap[def.LOCAL_ID].Dir == IO.MD_Down || currentMap[def.LOCAL_ID].Dir == IO.MD_Stop {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] = NO_ORDER
	}
	if currentMap[def.LOCAL_ID].Dir == IO.MD_Down && (!IsOrderBelow(currentMap) || currentFloor == 0) {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallUp] = NO_ORDER
	} else if currentMap[def.LOCAL_ID].Dir == IO.MD_Up && (!IsOrderAbove(currentMap) || currentFloor == def.NUMFLOORS - 1) {
		currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_HallDown] = NO_ORDER
	}

	currentMap[def.LOCAL_ID].Orders[currentFloor][IO.BT_Cab] = NO_ORDER
	SetElevMap(currentMap)

	return currentMap
}
/*
func GetNewEvent(newMap ElevatorMap) (ElevatorMap, [][]int) {
	currentMap := GetElevMap()
	var buttonChanges [][]int

	for elev := 0; elev < def.NUMELEVATORS; elev++ {
		if currentMap[def.LOCAL_ID].State != def.S_Dead && currentMap[def.LOCAL_ID].State != def.S_Init && newMap[def.LOCAL_ID].State != def.S_Dead && newMap[def.LOCAL_ID].State != def.S_Init {

			for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES; button++ {

					if newMap[elev].Orders[floor][button] == ORDER && currentMap[elev].Orders[floor][button] != ORDER {
						if button != IO.BT_Cab {
							currentMap[elev].Orders[floor][button] = newMap[elev].Orders[floor][button]
							buttonChanges = append(buttonChanges, []int{floor, button})
						} else {
							currentMap[elev].Orders[floor][button] = newMap[elev].Orders[floor][button]
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
*/

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
				emptyMap[elev].Orders[floor][button] = NO_ORDER
			}
		}
		emptyMap[def.LOCAL_ID].State = def.S_Dead
		emptyMap[def.LOCAL_ID].Active = true
		emptyMap[def.LOCAL_ID].Dir = IO.MD_Stop
		emptyMap[def.LOCAL_ID].Floor = -1
	}
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
