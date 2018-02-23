package def



const (
	//elevator constants
	NumFloors = 4
	NumElevators = 3
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ElevState int
const (
	DEAD ElevState 		= 0
	IDLE 				= 1
	MOVING 				= 2
	DOOROPEN 			= 3
	INIT  				= 4
)
