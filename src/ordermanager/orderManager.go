package ordermanager

import (
    "fmt"
    "../def"
    "sync"
    "time"
)

var mapMtx = &sync.Mutex{}
var localElevMap *ElevatorMap

//structs and constants
type ElevatorMap [def.NUMELEVATORS]def.Elev
/*
type CompleteElevMap struct {
    elevMap[NumElevators] elevatorMap
    allRequests ElevRequests
}

type BroadcastElevMap struct {
    elevMap elevatorMap
    requests ElevRequests

}
*/

type Request int
const (
    NO_ORDER Request         = 0
    ORDER                    = 1
    ORDER_ACCEPTED           = 2
    ORDER_IMPOSSIBLE         = -1
)

type ElevRequests []Request


//functions
func InitElevMap(backup bool) {
    fmt.Println("func: InitElevMap")
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

func PrintElevMap(){
  for{
      time.Sleep(time.Second)
      for elev := 0; elev < def.NUMELEVATORS; elev++ {
          fmt.Println("-----------------------------------------------")
          fmt.Println("Elevator number: ", localElevMap[elev].ElevID)
        switch localElevMap[elev].Dir {
        case -1:
          fmt.Println("Motor direction: Down")
        case 0:
          fmt.Println("Motor direction: Stop")
        case 1:
          fmt.Println("Motor direction: Up")
        }
          fmt.Println("Floor: ", localElevMap[elev].Floor)
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
        fmt.Println("Buttons: ", localElevMap[elev].Buttons)
        fmt.Println("Orders: ", localElevMap[elev].Orders)
      }
  }
}


func UpdateElevMap(newMap ElevatorMap) (ElevatorMap, bool){
    fmt.Println("UpdateElevMap")
    currentMap := GetElevMap()
    allChangesMade := false

    //update direction
    if newMap[def.LOCAL_ID].Dir != currentMap[def.LOCAL_ID].Dir {
      currentMap[def.LOCAL_ID].Dir = newMap[def.LOCAL_ID].Dir
      allChangesMade = true
    }

    //update floor
    if newMap[def.LOCAL_ID].Floor != currentMap[def.LOCAL_ID].Floor {
      currentMap[def.LOCAL_ID].Floor = newMap[def.LOCAL_ID].Floor
      allChangesMade = true
    }

    //update state
    if newMap[def.LOCAL_ID].State != currentMap[def.LOCAL_ID].State {
      currentMap[def.LOCAL_ID].State = newMap[def.LOCAL_ID].State
      allChangesMade = true
    }

    //update buttons and orders
    for elev := 0; elev < def.NUMELEVATORS; elev++ {
      if currentMap[def.LOCAL_ID].State != def.S_Dead && newMap[def.LOCAL_ID].State != def.S_Dead {

        for floor := 0; floor < def.NUMFLOORS; floor++ {
          for button := 0; button < def.NUMBUTTON_TYPES; button++ {

            //set buttons to 0 or 2 depending on if there is an order or not
            //if there is an accepted order, then buttons should be set to 2
            if newMap[elev].Orders[floor][button] == 1 && currentMap[elev].Buttons[floor][button] != 2 {
              if button != def.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = 2
                }
                allChangesMade = true
              }
            //if there was an order, but it has been completed, buttons should be set to 0 again
            } else if newMap[elev].Orders[floor][button] == 0 && currentMap[elev].Orders[floor][button] == 1 {
              if button != def.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = 0
                }
                allChangesMade = true
              }
            }


            //set all values to 1
            if newMap[elev].Buttons[floor][button] == 1 && currentMap[elev].Buttons[floor][button] == 0 {
              if button != def.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = newMap[e].Buttons[floor][button]
                }
                allChangesMade = true
              } else if elev == def.LOCAL_ID {
                currentMap[elev].Buttons[floor][def.BT_Cab] = newMap[elev].Buttons[floor][def.BT_Cab]
                allChangesMade = true
              }

            } else if newMap[elev].Buttons[floor][button] == 2 && currentMap[elev].Buttons[floor][button] != 2 {
              if button != def.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = newMap[e].Buttons[floor][button]
                }
                allChangesMade = true
              } else if elev == def.LOCAL_ID {
                currentMap[elev].Buttons[floor][def.BT_Cab] = newMap[elev].Buttons[floor][def.BT_Cab]
                allChangesMade = true
              }

            } else if newMap[elev].Buttons[floor][button] == 0 && currentMap[elev].Buttons[floor][button] == 2 {
              if button != def.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = newMap[e].Buttons[floor][button]
                }
                allChangesMade = true
              } else if elev == def.LOCAL_ID {
                currentMap[elev].Buttons[floor][def.BT_Cab] = newMap[elev].Buttons[floor][def.BT_Cab]
                allChangesMade = true
              }
            }
          }
        }
      }
    }

    MakeBackup(currentMap)
    SetElevMap(currentMap)

    return currentMap, allChangesMade
}


func GetNewEvent(newMap ElevatorMap) (ElevatorMap, [][]int) {
  fmt.Println("func: GetNewEvent")
  currentMap := GetElevMap()
  var buttonChanges [][]int

  for elev := 0; elev < def.NUMELEVATORS; elev++ {
    if currentMap[def.LOCAL_ID].State != def.S_Dead && newMap[def.LOCAL_ID].State != def.S_Dead {

      for floor := 0; floor < def.NUMFLOORS; floor++ {
        for button := 0; button < def.NUMBUTTON_TYPES; button++ {

          if newMap[elev].Buttons[floor][button] == 1 && currentMap[elev].Buttons[floor][button] != 1 {
            if button != def.BT_Cab {
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
    fmt.Println("func: SetElevMap")
    mapMtx.Lock()
    *localElevMap = newMap
    mapMtx.Unlock()
}


func GetElevMap() ElevatorMap {
    fmt.Println("func: GetElevMap")
    mapMtx.Lock()
    elevMap := *localElevMap
    mapMtx.Unlock()
    return elevMap
}


func MakeEmptyElevMap() *ElevatorMap {
    fmt.Println("func: MakeEmptyElevMap")
    emptyMap := new(ElevatorMap)

    for elev := 0; elev < def.NUMELEVATORS; elev++ {
        emptyMap[elev].ElevID = elev
        for floor := 0; floor < def.NUMFLOORS; floor++ {
            for button := 0; button < def.NUMBUTTON_TYPES; button++ {
                emptyMap[elev].Buttons[floor][button] = 0
                emptyMap[elev].Orders[floor][button] = 0
            }
        }
        emptyMap[def.LOCAL_ID].State = def.S_Idle
        emptyMap[def.LOCAL_ID].Dir = def.MD_Stop
        emptyMap[def.LOCAL_ID].Floor = -1
    }
    fmt.Println("Adresse til kartet: ", localElevMap)
    return emptyMap
}


func isClosestElevator() bool {
  fmt.Println("func: isClosestElevator")
  return false
}
