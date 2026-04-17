package db

import (
	"encoding/json"
	"os"
	"path/filepath"

	"dailytrack/models"
)

func GetConfigPath() (string, error) {
	workspace, err := GetWorkspacePath()
	if err != nil {
		return "", err
	}
	if workspace == "" {
		return "", os.ErrNotExist // Or a custom error indicating it's not setup yet
	}

	return filepath.Join(workspace, "config.json"), nil
}

func LoadConfig() (*models.Config, error) {
	path, err := GetConfigPath()
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var cfg models.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *models.Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
