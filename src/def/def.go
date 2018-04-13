package def

//const parameters
const (
	NUMFLOORS       = 4
	NUMELEVATORS    = 3
	NUMBUTTON_TYPES = 3

	ELEVATOR_DEAD = 0
	FLOOR_ARRIVAL = 1
	BUTTON_PUSHED = 2
	DOOR_TIMEOUT  = 3

	DOOR_TIMEOUT_TIME = 3
	IDLE_TIMEOUT_TIME = 1

	BACKUP_IP   = "	127.0.0.1:30000" //to be decided
	BACKUP_PORT = ":30000"
	SEND_MAP_PORT	= 30011	// Must be in range 30010 to 30100
	SEND_ID_PORT = 20011	// Must be in range 20010 to 20100
	LOCAL_ID    = 0				// Must be in rage 0 to NUMELEVATORS
)

type ElevState int

const (
	S_Dead     ElevState = 0
	S_Init               = 1
	S_Idle               = 2
	S_Moving             = 3
	S_DoorOpen           = 4
)

//structs
type NewEvent struct {
	EventType int
	Type      interface{}
}

type MapMessage struct {
	SendMap   interface{}
	SendEvent interface{}
}

func MakeMapMessage(elevmap interface{}, event interface{}) MapMessage {
	sendMessage := MapMessage{
		SendMap:   elevmap,
		SendEvent: event,
	}
	return sendMessage
}
