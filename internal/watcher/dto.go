package watcher

import "task-killer/internal/log"

type WindowsProcess struct {
	ProcessID       uint32
	ParentProcessID uint32
	FullPath        string
	Name            string
}

type WatcherInit struct {
	Log     *log.Logger
	IsDebug bool
}
