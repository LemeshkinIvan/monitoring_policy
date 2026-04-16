package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type TypeConnection string

const (
	Local TypeConnection = "local"
	SMB   TypeConnection = "smb"
	HTTP  TypeConnection = "http"
)

type CMDArgs struct {
	IsDebug       bool
	EnableLogFile bool
	ConfigPath    string
	HostAddress   string
	Conn          TypeConnection
}

// run main.exe -"192.168.100.110:3000\\distr\\task\\cfg.json" -"debug" -"enableFile"
func getCMDFlags() (CMDArgs, error) {
	var isDebug bool = false
	var enableLogFile bool = false

	// addr path
	confPath, err := getStringFlag("host", "", "config address")
	if err != nil {
		return CMDArgs{}, fmt.Errorf("confPath is nil")
	}

	addr, path := parseConfigPath(confPath)
	if addr == "" || path == "" {
		return CMDArgs{}, fmt.Errorf("addr or path is empty")
	}

	// enable type connection to config folder
	typeConn, err := getStringFlag("type_conn", "smb", "")
	if err != nil {
		return CMDArgs{}, fmt.Errorf("%w is nil", err)
	}

	isDebug, err = getBoolFlag("debug", "true", "")
	if err != nil {
		return CMDArgs{}, fmt.Errorf("%w is nil", err)
	}

	enableLogFile, err = getBoolFlag("logToFile", "false", "")
	if err != nil {
		return CMDArgs{}, fmt.Errorf("%w is nil", err)
	}

	return CMDArgs{
		HostAddress:   addr,
		IsDebug:       isDebug,
		ConfigPath:    path,
		Conn:          TypeConnection(typeConn),
		EnableLogFile: enableLogFile,
	}, nil
}

// func validateURL

func getStringFlag(name string, def string, description string) (string, error) {
	value := flag.String(name, def, description)
	if value == nil {
		return "", fmt.Errorf("%s is nil", name)
	}

	return *value, nil
}

func getBoolFlag(name string, def string, description string) (bool, error) {
	raw := flag.String(name, def, description)
	if raw == nil {
		return false, fmt.Errorf("%s is nil", name)
	}

	value, err := strconv.ParseBool(*raw)
	if err != nil {
		return false, err
	}

	return value, nil
}

func parseConfigPath(arg string) (string, string) {
	lines := strings.SplitN(arg, "\\", 2)

	sub1 := lines[0]
	sub2 := lines[1]

	return sub1, sub2
}
