package ordermanager

import (
    "fmt"
    "../network"
    "../def/"
)

type elevatorMap struct {
    dir MotorDirection
    IP string
    state ElevState
    orders

}



type completeElevMap struct {
    elevMap[NumElevators] elevatorMap
    allRequests ElevRequests
}

type broadcastElevMap struct {
    elevMap elevatorMap
    requests ElevRequests

}

type request int
const (
    NO_ORDER request         = 0
    ORDER                    = 1
    ORDER_ACCEPTED           = 2
)
type ElevRequests struct {
    requests []
}
