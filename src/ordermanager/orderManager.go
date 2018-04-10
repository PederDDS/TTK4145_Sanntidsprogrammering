package ordermanager

import (
    "fmt"
    "../def"
    "sync"
    "../IO"
    "math"
)

var mapMtx = &sync.Mutex{}
var localElevMap *ElevatorMap

//structs and constants
type ElevatorMap [def.NUMELEVATORS]Elev

type Elev struct {
	ElevID 		int
	Dir 		  IO.MotorDirection
	Floor 		int
	State 		def.ElevState
	Buttons 	[def.NUMFLOORS][def.NUMBUTTON_TYPES]int
	Orders 		[def.NUMFLOORS][def.NUMBUTTON_TYPES]int
}
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
      // Husk å endre 1 tilbake til def.NUMELEVATORS!!
      for elev := 0; elev < 1; elev++ {
          fmt.Println("-----------------------------------------------")
          fmt.Println("Elevator number:", localElevMap[elev].ElevID)

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
          fmt.Print("Floor: ", localElevMap[elev].Floor + 1, "th\n")
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


func UpdateElevMap(newMap ElevatorMap) (ElevatorMap, bool){
    fmt.Println("func: UpdateElevMap")
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
              if button != IO.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = 2
                }
                allChangesMade = true
              }
            //if there was an order, but it has been completed, buttons should be set to 0 again
            } else if newMap[elev].Orders[floor][button] == 0 && currentMap[elev].Orders[floor][button] == 1 {
              if button != IO.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = 0
                }
                allChangesMade = true
              }
            }


            //set all values to 1
            if newMap[elev].Buttons[floor][button] == 1 && currentMap[elev].Buttons[floor][button] == 0 {
              if button != IO.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = newMap[e].Buttons[floor][button]
                }
                allChangesMade = true
              } else if elev == def.LOCAL_ID {
                currentMap[elev].Buttons[floor][IO.BT_Cab] = newMap[elev].Buttons[floor][IO.BT_Cab]
                allChangesMade = true
              }

            } else if newMap[elev].Buttons[floor][button] == 2 && currentMap[elev].Buttons[floor][button] != 2 {
              if button != IO.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = newMap[e].Buttons[floor][button]
                }
                allChangesMade = true
              } else if elev == def.LOCAL_ID {
                currentMap[elev].Buttons[floor][IO.BT_Cab] = newMap[elev].Buttons[floor][IO.BT_Cab]
                allChangesMade = true
              }

            } else if newMap[elev].Buttons[floor][button] == 0 && currentMap[elev].Buttons[floor][button] == 2 {
              if button != IO.BT_Cab {
                for e := 0; e < def.NUMELEVATORS; e++ {
                  currentMap[e].Buttons[floor][button] = newMap[e].Buttons[floor][button]
                }
                allChangesMade = true
              } else if elev == def.LOCAL_ID {
                currentMap[elev].Buttons[floor][IO.BT_Cab] = newMap[elev].Buttons[floor][IO.BT_Cab]
                allChangesMade = true
              }
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


func GetNewEvent(newMap ElevatorMap) (ElevatorMap, [][]int) {
  fmt.Println("func: GetNewEvent")
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
                emptyMap[elev].Orders[floor][button] = ORDER_IMPOSSIBLE
            }
        }
        emptyMap[def.LOCAL_ID].State = def.S_Dead
        emptyMap[def.LOCAL_ID].Dir = IO.MD_Stop
        emptyMap[def.LOCAL_ID].Floor = -1
    }
    PrintElevMap()
    return emptyMap
}


func IsClosestElevator(currentMap ElevatorMap, floor int) bool {
  fmt.Println("func: IsClosestElevator")
  result := true
  myDistance := int(math.Abs(float64(currentMap[def.LOCAL_ID].Floor - floor)))

  if currentMap[def.LOCAL_ID].Floor < floor {

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
