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

	DOOR_TIMEOUT_TIME = 1
	IDLE_TIMEOUT_TIME = 10

	BACKUP_IP   = "	127.0.0.1:30000" //to be decided
	BACKUP_PORT = ":30000"
	LOCAL_ID    = 0
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
