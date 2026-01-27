package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type GlobalConfig struct {
	Language      string   `json:"language"`       // "es" or "en"
	SkillsSources []string `json:"skills_sources"` // Global default sources
}

func GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kolyn", "config.json"), nil
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing global config: %w", err)
	}
	return &cfg, nil
}

func SaveGlobalConfig(cfg *GlobalConfig) error {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
