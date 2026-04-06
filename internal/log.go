package internal

import (
	"fmt"
	"os"
	"time"
)

const mask = "02-01-2006 15:04:05"
const defaultLogFolder = "C:\\Users\\%s\\AppData\\Local\\blacklist\\logs"
const deafultLogFileName = "app.log"

var logChan = make(chan string, 100) // буфер!
var name = os.Getenv("USERNAME")

func Log(rawMsg string) {
	msg := formatMsg(rawMsg)

	select {
	case logChan <- msg:
	default:
		// stdout
		fmt.Print(msg)
	}
}

func LogStdOut(rawMsg string) {
	fmt.Println(formatMsg(rawMsg))
}

func formatMsg(raw string) string {
	var msg string
	msg = time.Now().Format(mask)

	if msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	msg += " " + raw
	return msg
}

func LogInFile(fileName string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for msg := range logChan {
		file.WriteString(msg + "\n")
	}
}

func CreateLogFolder(path string) error {
	if len(path) == 0 {
		path = fmt.Sprintf(defaultLogFolder, name)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func CreateLogFile() error {
	_, err := os.Create(deafultLogFileName + "\\" + deafultLogFileName)
	if err != nil {
		return err
	}
	return nil
}
