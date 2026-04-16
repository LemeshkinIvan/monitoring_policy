package main

import (
	"errors"
	"fmt"
	"os"

	cfg "task-killer/internal/config"
	providers "task-killer/internal/config_providers"
	"task-killer/internal/constants"
	"task-killer/internal/log"
	w "task-killer/internal/watcher"
	"time"
)

const (
	defaultTimeIdle    = 10 * time.Second // сколько ждать при завершении интерации watcher
	defaultTimeRequest = 2 * time.Second  // при истечении пойдет за файлом
)

func main() {
	// init
	args, err := getCMDFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	logger, err := log.NewLogger(log.LoggerCfg{
		IsDebug:        args.IsDebug,
		EnableWriteLog: args.EnableLogFile,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// close goroutine
	if args.EnableLogFile {
		defer logger.Close()
	}

	logger.Log(constants.StartProgram, log.INFO)

	var watcher *w.Win32Watcher
	var config *cfg.ConfigDTO

	smb, err := providers.NewSMBManager(providers.SMBInput{
		Addr: args.HostAddress,
	})
	if err != nil {
		logger.Log(err.Error(), log.FATAL)
		os.Exit(-1)
	}

	var confManager *cfg.ConfigManager = &cfg.ConfigManager{
		SMBClient: smb,
	}

	watcher, err = w.NewWin32Watcher(w.WatcherInit{
		Log:     logger,
		IsDebug: args.IsDebug,
	})
	if err != nil {
		logger.Log(err.Error(), log.FATAL)
		os.Exit(-1)
	}

	// start
	for {
		logger.Log(constants.GetConfig, log.INFO)

		// лезем за конфигом
		for config == nil {
			config, err = confManager.GetConfigWithSMB(args.ConfigPath)

			if err != nil {
				logger.Log(err.Error(), log.WARN)
			}

			time.Sleep(defaultTimeRequest)
		}

		logger.Log(constants.ConfigIsLoaded, log.INFO)

		sleepDur, err := time.ParseDuration(config.TimeSleep)
		if err != nil {
			sleepDur = defaultTimeIdle
			logger.Log(err.Error(), log.WARN)
			logger.Log(constants.SetDefaultSleepTime, log.WARN)
		}

		if err := watcher.StartWatcherWin32(config.Blacklist); err != nil {
			if errors.Is(err, w.ErrBlacklistLen) {
				logger.Log(err.Error(), log.WARN)
			} else {
				logger.Log(err.Error(), log.FATAL)
				os.Exit(-1)
			}
		}

		logger.Log(constants.GetSleepingMsg(config.TimeSleep), log.INFO)
		time.Sleep(sleepDur)
	}
}
