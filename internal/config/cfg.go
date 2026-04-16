package cfg

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	prov "task-killer/internal/config_providers"
)

type ConfigDTO struct {
	Blacklist   []string `json:"blacklist"`
	TimeRequest string   `json:"time_cfg_request"`
	TimeSleep   string   `json:"time_sleep"`
	LogPath     string   `json:"log_path"`
}

type ConfigManager struct {
	SMBClient *prov.SMBManager
	// local
	// http
}

func (c *ConfigManager) GetConfigWithSMB(path string) (*ConfigDTO, error) {
	var clonedFilePath string = "config.json"

	if err := c.SMBClient.ConnectToShare(); err != nil {
		return nil, err
	}

	defer c.SMBClient.Disconnect()

	share := c.SMBClient.GetShare()

	src, err := share.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(clonedFilePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, err
	}

	fmt.Println("Downloaded!")

	file, err := os.ReadFile(clonedFilePath)
	if err != nil {
		return nil, err
	}

	var cfg ConfigDTO
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
