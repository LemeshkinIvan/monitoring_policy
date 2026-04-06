package main

import (
	"fmt"
	"os"
	"task-killer/internal"
	"time"
)

const (
	path               = "../cfg/cfg.json"
	devPath            = "../cfg/cfg.json"
	defaultTimeIdle    = 10 * time.Second // сколько ждать если getConfig дал ошибку
	defaultTimeRequest = 30 * time.Second // при истечении пойдет за файлом
)

// flow
// программа по вшитому пути спрашивает cfg.json. парсит в структуру
// если getConfig вернула nil вместо структуры, то программа отлогирует ошибку
// и после небольшого таймаута пойдет в следующую итерацию цикла, чтобы таки запросить cfg
// если json валиден, то после идет валидация полей структуры.
// если что-то отваливается, к примеру, timeout'ы, то ставим default

func main() {
	//go internal.LogInFile("app.log")

	internal.LogStdOut("start program")

	for {
		var cfg *internal.ConfigDTO
		var err error

		internal.LogStdOut("start getConfig")
		for cfg == nil {
			cfg, err = internal.GetConfig(devPath)
			if err != nil {
				internal.LogStdOut(err.Error())
			}

			time.Sleep(defaultTimeRequest)
		}

		internal.LogStdOut("yep, i get it!")

		// ставим время сна цикла
		sleepDur, err := time.ParseDuration(cfg.TimeIdle)
		if err != nil {
			sleepDur = defaultTimeIdle
			internal.LogStdOut(err.Error())
			internal.LogStdOut("set default time idle")
		}

		if err := initLog(); err != nil {
			internal.LogStdOut(err.Error())
			os.Exit(-1)
		}

		if err := internal.StartWatcherWin32(cfg); err != nil {
			internal.LogStdOut(err.Error())
			os.Exit(-1)
		}

		internal.LogStdOut(fmt.Sprintf("okey, start sleeping. Dur: %s", cfg.TimeIdle))
		time.Sleep(sleepDur)
	}
}

func initLog() error {
	if err := internal.CreateLogFolder(""); err != nil {
		internal.LogStdOut("i cant create log folder. please fix it. Bye")
		return fmt.Errorf(err.Error())
	}

	if err := internal.CreateLogFile(); err != nil {
		internal.LogStdOut("i cant create log file. please fix it. Bye")
		return fmt.Errorf(err.Error())
	}

	return nil
}
