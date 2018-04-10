package IO


import(
  "sync"
  "../def"
  "time"
)

var mutex = &sync.Mutex{}

func initHW(msg_toHW chan def.MapMessage, msg_fromHWFloor chan def.MapMessage, msg_fromHWButton chan def.MapMessage) {
  go PollFloorEvents(msg_fromHWFloor)
  go PollButtonEvents(msg_fromHWButton)
  go SetLights(msg_toHW)
}

func PollButtonEvents(msg_fromHWButton chan def.MapMessage) {

}


func PollFloorEvents(msg_fromHWFloor chan def.MapMessage) {

}


func SetLights(msg_toHW chan def.MapMessage) {
  for {
    select {
    case msg := <-msg_toHW:
      currentMap := msg.SendMap.(ordermanager.ElevatorMap)

      for button := 0; button < def.NUMBUTTON_TYPES; button++ {
        for floor := 0; floor < def.NUMFLOORS; floor++ {
          setLight := 1

          if button != BT_Cab {
            for elev := 0; elev < def.NUMELEVATORS; elev++ {
              if currentMap[elev].Buttons[floor][button] != 1 && currentMap[elev].State != def.S_Dead {
                setLight = 0
              }
            }

          } else if currentMap[def.LOCAL_ID].Buttons[floor][button] != 1 {
            setLight = 0
          }

          if button == 0 {
            setButton := BT_HallUp
          } else if button == 1 {
            setButton := BT_HallDown
          } else {
            setButton := BT_Cab
          }
          SetButtonLamp(setButton, floor, setLight)
        }
      }
      SetFloorIndicator(currentMap[def.LOCAL_ID].Floor)
    }
    time.Sleep(20*time.Millisecond)
  }
}
