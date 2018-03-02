package ordermanager

import (
    "fmt"
    "../network"
    "../def/"
)

var mapMtx sync.Mutex
var localElevMap *ElevatorMap

//structs and constants
type ElevatorMap[def.NUMELEVATORS]def.Elev struct
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
    NO_ORDER request         = 0
    ORDER                    = 1
    ORDER_ACCEPTED           = 2
    ORDER_IMPOSSIBLE         = -1
)

type ElevRequests struct {
    requests []
}


//functions
func InitElevMap() {
    mapMtx.Lock()
    localElevMap := new(ElevatorMap)

    makeBackup(*localElevMap)
    mapMtx.Unlock()
}


func UpdateElevMap(newMap ElevatorMap, userID int) (ElevatorMap, bool){
    currentMap := GetElevMap()



    makeBackup(UpdateMap)
    SetElevMap(updatedMap)
    return updatedMap, boolVal
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

    for el int; el<NUMELEVATORS; el++ {
        emptyMap.ElevID = el
        for fl int; fl<NUMFLOORS; fl++ {
            for b int; b<NUMBUTTON_TYPES; b++ {
                emptyMap[el].Buttons[fl][b] = 0
                emptyMap[el].Orders[fl][b] = 0
            }
        }
        emptyMap.State = def.S_Idle
        emptyMap.Dir = def.MD_Stop
        emptyMap.Floor = -1 //kanskje 0 i stedet?
    }
    return emptyMap
}
