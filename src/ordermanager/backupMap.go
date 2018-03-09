package ordermanager

import (
	"../def"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func SoftwareBackup() {
	backupTicker := time.NewTicker(250 * time.Millisecond)
	newBackup := exec.Command("gnome-terminal", "-x", "sh", "-c", "make run")
	err := newBackup.Run()
	if err != nil {
		fmt.Println("Unable to spawn backup. You're on your own!")
		return
	}

	backupAdr, err := net.ResolveUDPAddr("udp", def.BACKUP_IP)
	if err != nil {
		return
	}

	backupConn, err := net.DialUDP("udp", nil, backupAdr)
	if err != nil {
		return
	}

	aliveMessage := true

	for {
		select {
		case <-backupTicker.C:
			jsonBuf, _ := json.Marshal(aliveMessage)
			backupConn.Write(jsonBuf)
		}

	}

}

func ReadBackup() ElevatorMap {
	backupFile, err := ioutil.ReadFile("src/ordermanager/backup.txt")

	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(strings.NewReader(string(backupFile)))
	csvReader.FieldsPerRecord = -1

	stringMap := [][]string{}

	for {
		csvLine, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		stringMap = append(stringMap, csvLine)
	}

	backupMap := *MakeEmptyElevMap()

	for elevator := 0; elevator < def.NUMELEVATORS; elevator++ {
		backupMap[elevator].ID, _ = strconv.Atoi(stringMap[e*(5+def.NUMFLOORS)][0])
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			for button := 0; button < def.NUMBUTTONS; button++ {
				backupMap[elevator].Buttons[floor][button], _ = strconv.Atoi(stringMap[elevator*(5+def.NUMFLOORS)+1+floor][button])
			}
		}
		backupMap[elevator].Direction, _ = strconv.Atoi(stringMap[elevator*(5+def.NUMFLOORS)+def.NUMFLOORS+1][0])
		backupMap[elevator].Position, _ = strconv.Atoi(stringMap[elevator*(5+def.NUMFLOORS)+def.NUMFLOORS+2][0])
		backupMap[elevator].Door, _ = strconv.Atoi(stringMap[elevator*(5+def.NUMFLOORS)+def.NUMFLOORS+3][0])
		backupMap[elevator].IsAlive, _ = strconv.Atoi(stringMap[elevator*(5+def.NUMFLOORS)+def.NUMFLOORS+4][0])
	}

	return backupMap

}


func WriteBackup(backupMap ElevatorMap) {
	backupFile, err := os.Create("src/ordermanager/backup.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer backupFile.Close()

	stringMap := [][]string{}

	for elevator := 0; elevator < def.NUMELEVATORS; elevator++ {
		stringMap = append(stringMap, []string{strconv.Itoa(backupMap[elevator].ID)})
		for floor := 0; floor < def.NUMFLOORS; floor++ {
			stringArray := []string{}
			for button := 0; button < def.NUMBUTTONS; button++ {
				stringArray = append(stringArray, strconv.Itoa(backupMap[elevator].Buttons[floors][buttons]))
			}
			stringMap = append(stringMap, stringArray)
		}
		stringMap = append(stringMap, []string{strconv.Itoa(backupMap[elevator].Direction)})
		stringMap = append(stringMap, []string{strconv.Itoa(backupMap[elevator].Position)})
		stringMap = append(stringMap, []string{strconv.Itoa(backupMap[elevator].Door)})
		stringMap = append(stringMap, []string{strconv.Itoa(backupMap[elevator].IsAlive)})
	}
	backupWriter := csv.NewWriter(backupFile)
	err = backupWriter.WriteAll(stringMap)
	if err != nil {
		log.Fatal(err)
	}
}
