package internal

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const TH32CS_SNAPPROCESS = 0x00000002

var protected = [5]string{"System32", "SystemApps"}

type WindowsProcess struct {
	ProcessID       uint32
	ParentProcessID uint32
	FullPath        string
	Name            string
}

func GetSnapshot() (windows.Handle, error) {
	return windows.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)
}

func CloseSnapshot(handle windows.Handle) error {
	return windows.CloseHandle(handle)
}

func StartWatcherWin32(cfg *ConfigDTO) error {
	snapshot, err := GetSnapshot()
	if err != nil {
		return err
	}

	defer CloseSnapshot(snapshot)

	var cursor windows.ProcessEntry32
	cursor.Size = uint32(unsafe.Sizeof(cursor))
	// get the first process
	err = windows.Process32First(snapshot, &cursor)
	if err != nil {
		return err
	}

	// append first
	allProcess := []WindowsProcess{}
	allProcess = append(allProcess, newWindowsProcess(&cursor))

	// append all next
	nextResult, err := getProcess(snapshot, cursor)
	if err != nil {
		return err
	}

	allProcess = append(allProcess, nextResult...)
	for _, i := range allProcess {
		// debug
		Log(fmt.Sprintf("found process with pid: %d, name: %s\npath: %s", i.ProcessID, i.Name, i.FullPath))

		if isProtected(i.FullPath) {
			LogStdOut("its system procces. dont touch")
			continue
		}

		for _, j := range cfg.Blacklist {
			if strings.EqualFold(i.Name, j) {
				if err := killProcess(i.ProcessID); err != nil {
					LogStdOut(err.Error())
					continue
				}

				LogStdOut(fmt.Sprintf("process with pid: %d, name: %s was killed", i.ProcessID, i.Name))
			}
		}
	}

	return nil
}

func isProtected(path string) bool {
	for _, i := range protected {
		if strings.Contains(path, i) {
			return true
		}
	}
	return false
}

func getProcess(snapshot windows.Handle, cursor windows.ProcessEntry32) ([]WindowsProcess, error) {
	results := make([]WindowsProcess, 0, 50)
	for {
		results = append(results, newWindowsProcess(&cursor))

		err := windows.Process32Next(snapshot, &cursor)
		if err != nil {
			// windows sends ERROR_NO_MORE_FILES on last process
			if err == syscall.ERROR_NO_MORE_FILES {
				return results, nil
			}
			return nil, err
		}
	}
}

func killProcess(pid uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_TERMINATE,
		false,
		pid,
	)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	err = windows.TerminateProcess(handle, 1)
	if err != nil {
		return err
	}

	return nil
}

func getProcessPath(pid uint32) (string, error) {
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

func newWindowsProcess(e *windows.ProcessEntry32) WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	path, err := getProcessPath(e.ProcessID)
	if err != nil {
		Log(err.Error())
	}

	return WindowsProcess{
		ProcessID:       e.ProcessID,
		ParentProcessID: e.ParentProcessID,
		Name:            syscall.UTF16ToString(e.ExeFile[:end]),
		FullPath:        path,
	}
}
