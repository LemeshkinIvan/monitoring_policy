package watcher

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"task-killer/internal/log"
	"unsafe"

	"golang.org/x/sys/windows"
)

const TH32CS_SNAPPROCESS = 0x00000002
const del = "\n----------------------------------------------\n"
const unknownProccessPath = "unknown"

type Win32Watcher struct {
	Logger    *log.Logger
	IsDebug   bool
	protected []string
}

func NewWin32Watcher(cfg WatcherInit) (*Win32Watcher, error) {
	if cfg.Log == nil {
		return nil, ErrLogNullable
	}

	return &Win32Watcher{
		Logger:  cfg.Log,
		IsDebug: cfg.IsDebug,
		protected: []string{
			"System32",
			"SystemApps",
		}}, nil
}

func (w *Win32Watcher) GetSnapshot() (windows.Handle, error) {
	return windows.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)
}

func (w *Win32Watcher) CloseSnapshot(handle windows.Handle) error {
	return windows.CloseHandle(handle)
}

func (w *Win32Watcher) StartWatcherWin32(blacklist []string) error {
	if len(blacklist) == 0 {
		return ErrBlacklistLen
	}

	snapshot, err := w.GetSnapshot()
	if err != nil {
		return err
	}

	defer w.CloseSnapshot(snapshot)

	var cursor windows.ProcessEntry32
	cursor.Size = uint32(unsafe.Sizeof(cursor))
	// get the first process
	err = windows.Process32First(snapshot, &cursor)
	if err != nil {
		return err
	}

	pr, err := w.newWindowsProcess(&cursor)
	if err != nil {
		return err
	}

	if pr == nil {
		return fmt.Errorf("procces is nil")
	}

	// append first
	allProcess := []WindowsProcess{}
	allProcess = append(allProcess, *pr)

	// append all next
	nextResult, err := w.getProcess(snapshot, cursor)
	if err != nil {
		return err
	}

	allProcess = append(allProcess, nextResult...)
	for n, i := range allProcess {
		if w.IsDebug {
			w.writeDelim(n)
			w.Logger.Log(fmt.Sprintf("pid: %d, name: %s", i.ProcessID, i.Name), log.INFO)
		}

		if w.isProtected(i.FullPath) {
			if w.IsDebug {
				w.Logger.Log("its system procces. dont touch", log.WARN)
			}
			continue
		}

		for _, j := range blacklist {
			if strings.EqualFold(i.Name, j) {
				if err := w.killProcess(i.ProcessID); err != nil {
					if w.IsDebug {
						w.Logger.Log(err.Error(), log.WARN)
					}
					continue
				}

				if w.IsDebug {
					w.Logger.Log(fmt.Sprintf("pid: %d, name: %s was killed", i.ProcessID, i.Name), log.INFO)
				}
			}
		}
	}

	return nil
}

func (w *Win32Watcher) writeDelim(ind int) {
	w.Logger.Log(fmt.Sprintf("%s№:%d", del, ind), log.INFO)
}

func (w *Win32Watcher) isProtected(path string) bool {
	if path == unknownProccessPath {
		return true
	}

	path = strings.ToLower(filepath.Clean(path))

	for _, p := range w.protected {
		if strings.Contains(path, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func (w *Win32Watcher) getProcess(snapshot windows.Handle, cursor windows.ProcessEntry32) ([]WindowsProcess, error) {
	results := make([]WindowsProcess, 0, 255)
	for {
		pr, err := w.newWindowsProcess(&cursor)
		if err != nil {
			return nil, err
		}

		if pr == nil {
			return nil, fmt.Errorf("procces is nil")
		}

		results = append(results, *pr)

		err = windows.Process32Next(snapshot, &cursor)
		if err != nil {
			// windows sends ERROR_NO_MORE_FILES on last process
			if err == syscall.ERROR_NO_MORE_FILES {
				return results, nil
			}
			return nil, err
		}
	}
}

func (w *Win32Watcher) killProcess(pid uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_TERMINATE,
		false,
		pid,
	)
	if err != nil {
		return err
	}

	if handle == 0 {
		return ErrHandle
	}

	defer windows.CloseHandle(handle)

	err = windows.TerminateProcess(handle, 1)
	if err != nil {
		return err
	}

	return nil
}

func (w *Win32Watcher) getProcessPath(pid uint32) (string, error) {
	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return "", err
	}

	defer windows.CloseHandle(handle)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))

	err = windows.QueryFullProcessImageName(
		handle,
		0,
		&buf[0],
		&size,
	)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(buf[:]), nil
}

func (w *Win32Watcher) newWindowsProcess(e *windows.ProcessEntry32) (*WindowsProcess, error) {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	path, err := w.getProcessPath(e.ProcessID)
	if err != nil {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			if errno == 5 { // Access denied
				path = unknownProccessPath
			}
		} else {
			return nil, err
		}
	}

	return &WindowsProcess{
		ProcessID:       e.ProcessID,
		ParentProcessID: e.ParentProcessID,
		Name:            syscall.UTF16ToString(e.ExeFile[:end]),
		FullPath:        path,
	}, nil
}
