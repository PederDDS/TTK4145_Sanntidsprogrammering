package def


//const parameters
const (
	NUMFLOORS 			= 4
	NUMELEVATORS 		= 3
	NUMBUTTON_TYPES		= 3

	ELEVATOR_DEAD 		= 0
	FLOOR_ARRIVAL 		= 1
	BUTTON_PUSHED 		= 2
	DOOR_TIMEOUT 		= 3

	BACKUP_IP 			= "	127.0.0.1:30000" //to be decided
	BACKUP_PORT 		= ":30000"
	LOCAL_ID 			= 0

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

type ElevState int
const (
	S_Dead ElevState 	= 0
	S_Init 						= 1
	S_Idle 						= 2
	S_Moving	 				= 3
	S_DoorOpen  			= 4
)




//structs
type ButtonEvent struct {
	Floor 	int
	Button 	ButtonType
}

type Elev struct {
	ElevID 		int
	Dir 		  MotorDirection
	Floor 		int
	State 		ElevState
	Buttons 	[NUMFLOORS][NUMBUTTON_TYPES]int
	Orders 		[NUMFLOORS][NUMBUTTON_TYPES]int
}

type MapMessage struct {
	sendMap 	interface{}
	sendEvent 	interface{}
}

func MakeMapMessage(elevmap interface{}, event interface{}) MapMessage {
	sendMessage := MapMessage{
		sendMap: elevmap,
		sendEvent: event,
	}
	return sendMessage
}
