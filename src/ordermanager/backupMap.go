package elevatorMap

import (
	"../def"
	//"fmt"
	"sync"
)

var mapMutex = &sync.Mutex{}
var localMap *ElevMap

func InitMap(backup bool) {
	mapMutex.Lock()
	localMap = new(ElevMap)
	if backup {
		*localMap = readBackup()
	} else {
		localMap = NewCleanElevMap()
	}

	writeBackup(*localMap)
	mapMutex.Unlock()

}

func AddNewMapChanges(receivedMap ElevMap, user int) (ElevMap, bool) {
	currentMap := GetLocalMap()
	changeMade := false
	floorWithDoorOpen := -1

	if receivedMap[def.MY_ID].Door != currentMap[def.MY_ID].Door {
		if receivedMap[def.MY_ID].Door != -1 {
			floorWithDoorOpen = receivedMap[def.MY_ID].Door
		}
		currentMap[def.MY_ID].Door = receivedMap[def.MY_ID].Door
		changeMade = true
	}

	if receivedMap[def.MY_ID].Direction != currentMap[def.MY_ID].Direction {
		currentMap[def.MY_ID].Direction = receivedMap[def.MY_ID].Direction
		changeMade = true
	}

	if receivedMap[def.MY_ID].Position != currentMap[def.MY_ID].Position {
		currentMap[def.MY_ID].Position = receivedMap[def.MY_ID].Position
		changeMade = true
	}

	for elelvator := 0; elelvator < def.NUMELEVATORS; elevator++ {
		if currentMap[elevator].IsAlive == true && receivedMap[elevator].IsAlive != true {
			currentMap[elevator].IsAlive = true
		}
		if currentMap[elevator].IsAlive == true {
      for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES; button++ {

					if receivedMap[elevator].Buttons[floor][button] == true && currentMap[elevator].Buttons[floor][button] != true {
						if button != def.BT_Cab {
							currentMap[elevator].Buttons[floor][button] = receivedMap[elevator].Buttons[floor][button]
							currentMap[def.MY_ID].Buttons[floor][button] = receivedMap[elevator].Buttons[floor][button]
							changeMade = true
						} else if elevator == def.MY_ID {
							currentMap[elevator].Buttons[floor][def.BT_Cab] = receivedMap[elevator].Buttons[floor][def.BT_Cab]
							changeMade = true
						}
					} else if receivedMap[elevator].Buttons[floor][button] == false && floorWithDoorOpen == floor {
						if button != def.BT_Cab {
							currentMap[elevator].Buttons[floor][button] = receivedMap[elevator].Buttons[floor][button]
							changeMade = true
						} else if elevator == def.MY_ID {
							currentMap[elevator].Buttons[floor][def.BT_Cab] = receivedMap[elevator].Buttons[floor][def.BT_Cab]
							changeMade = true
						}
					}
				}
			}
		}

	}

	setLocalMap(currentMap)
	writeBackup(currentMap)
	return currentMap, changeMade
}

func GetEventFromNetwork(receivedMap ElevMap) ([][]int, ElevMap) {
	currentMap := GetLocalMap()
	floorWithDoorOpen := -1
	var newButtonPushes [][]int

	for elevator := 0; elevator < def.NUMELEVATORS; elevator++ {
		if receivedMap[elevator].Door != -1 {
			floorWithDoorOpen = receivedMap[elevator].Door
		}
	}

	for elevator := 0; elevator < def.NUMELEVATORS; elevator++ {
		if currentMap[elevator].IsAlive != true {
			currentMap[elevator].IsAlive = true
		}
		if receivedMap[elevator].IsAlive == true {
			for floor := 0; floor < def.NUMFLOORS; floor++ {
				for button := 0; button < def.NUMBUTTON_TYPES; button++ {

					if receivedMap[elevator].Buttons[floor][button] == true && currentMap[elevator].Buttons[floor][button] != true {
						if button != def.BT_Cab {
							currentMap[elevator].Buttons[floor][button] = receivedMap[elevator].Buttons[floor][button]
							newButtonPushes = append(newButtonPushes, []int{floor, button})

						} else {
							currentMap[elevator].Buttons[floor][button] = receivedMap[elevator].Buttons[floor][button]
						}
					} else if receivedMap[elevator].Buttons[floor][button] == false && floorWithDoorOpen == floor {
						currentMap[elevator].Buttons[floor][button] = receivedMap[elevator].Buttons[floor][button]
					}
				}
			}

			if receivedMap[elevator].Direction != currentMap[elevator].Direction && elevator != def.MY_ID {
				currentMap[elevator].Direction = receivedMap[elevator].Direction
			}
			if receivedMap[elevator].Position != currentMap[elevator].Position && elevator != def.MY_ID {
				currentMap[elevator].Position = receivedMap[elevator].Position
			}

		}
	}

	setLocalMap(currentMap)
	writeBackup(currentMap)
	return newButtonPushes, currentMap
}

func GetLocalMap() ElevMap {
	mapMutex.Lock()
	currentMap := *localMap
	mapMutex.Unlock()
	return currentMap
}

type ElevMap [def.NUMELEVATORS]def.ElevatorInfo

func NewCleanElevMap() *ElevMap {

	newMap := new(ElevMap)

	for elevator := 0; elevator < def.NUMELEVATORS; elevator++ {
		newMap[elevator].ID = elevator
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			for button := 0; button < def.NUMBUTTON_TYPES; button++ {
				newMap[elevator].Buttons[floor][button] = 0
			}
		}
		newMap[elevator].Direction = def.STILL
		newMap[elevator].Position = 0
		newMap[elevator].Door = -1
		newMap[elevator].IsAlive = true
	}
	return newMap
}

func setLocalMap(newMap ElevMap) {
	mapMutex.Lock()
	*localMap = newMap
	mapMutex.Unlock()
}
