package ordermanager

import (
	//"../def"
	//"fmt"
	//"sync"
	//"log"
)

func MakeBackup(backup ElevatorMap) {
	i := 0
}


func GetBackup() ElevatorMap {
	i := 0
}


func Backup() {
	newBackup := exec.Command("gnome-terminal", "-x", "-c", "sh", "en eller annen run command")
	err := newBackup.Run()
	if err != nil {
		log.Fatal(err)
	}


}
