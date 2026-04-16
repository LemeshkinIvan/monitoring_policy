package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel string

const INFO LogLevel = "INFO"
const WARN LogLevel = "WARN"
const FATAL LogLevel = "FATAL"

type Color string

const BLUE Color = "\033[34m"
const YELLOW Color = "\033[33m"
const RED Color = "\033[31m"

const Close string = "\033[0m"

type Logger struct {
	timeMask       string
	LogFolder      string
	NameFile       string
	IsDebug        bool
	EnableWriteLog bool

	Writer io.Writer
	Ch     chan string
}

type LoggerCfg struct {
	LogFolder      string
	NameFile       string
	IsDebug        bool
	EnableWriteLog bool
}

func NewLogger(cfg LoggerCfg) (*Logger, error) {
	if cfg.NameFile == "" {
		cfg.NameFile = "app.log"
	}

	l := &Logger{
		timeMask:       "02-01-2006 15:04:05",
		NameFile:       cfg.NameFile,
		IsDebug:        cfg.IsDebug,
		EnableWriteLog: cfg.EnableWriteLog,
	}

	if cfg.EnableWriteLog {
		// create folder and open file
		if cfg.LogFolder == "" {
			path, err := setDefPath()
			if err != nil {
				return nil, err
			}

			cfg.LogFolder = path
		}

		l.LogFolder = cfg.LogFolder

		err := l.InitLogFolder()
		if err != nil {
			return nil, err
		}

		l.Writer = &lumberjack.Logger{
			Filename:   l.LogFolder + "\\" + l.NameFile,
			MaxSize:    10, // MB
			MaxBackups: 5,
			MaxAge:     7,    // дней
			Compress:   true, // gzip
		}
	}

	l.Ch = make(chan string, 100)
	l.listen()

	return l, nil
}

func (l *Logger) listen() {
	for msg := range l.Ch {
		if l.Writer == nil {
			continue
		}

		_, err := l.Writer.Write([]byte(msg + "\n"))
		if err != nil {
			l.WriteToStdOut(err.Error(), string(WARN))
		}
	}
}

func (l *Logger) Close() {
	if l.EnableWriteLog {
		close(l.Ch)
	}
}

func (l *Logger) InitLogFolder() error {
	var path string
	var name = os.Getenv("USERNAME")

	// create folder
	if len(l.LogFolder) == 0 {
		path = fmt.Sprintf(l.LogFolder, name)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (l *Logger) Log(msg string, level LogLevel) {
	if level == "" {
		level = INFO
	}

	var header string
	switch level {
	case INFO:
		header = string(BLUE) + string(INFO) + Close
	case WARN:
		header = string(YELLOW) + string(WARN) + Close
	case FATAL:
		header = string(RED) + string(FATAL) + Close
	default:
		header = string(BLUE) + string(INFO) + Close
	}

	if l.IsDebug {
		l.WriteToStdOut(msg, header)
	}

	if l.EnableWriteLog {
		select {
		case l.Ch <- l.formatMsg(msg, header):
		default:
			// канал переполнен — не блокируемся
			l.WriteToStdOut("log channel overflow", string(YELLOW)+string(WARN)+Close)
		}
	}
}

func (l *Logger) WriteToStdOut(msg string, header string) {
	fmt.Println(l.formatMsg(msg, header))
}

func (l *Logger) formatMsg(raw string, header string) string {
	var msg string
	msg = time.Now().Format(l.timeMask) + " " + header

	if msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	msg += " " + raw
	return msg
}
