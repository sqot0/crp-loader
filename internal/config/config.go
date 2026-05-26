package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Config struct {
	ServerURL string `json:"server_url"`
}

func Load() (*Config, error) {
	executable, err := os.Executable()
	if err != nil {
		fmt.Println("Ошибка определения пути к исполняемому файлу:", err)
		return nil, err
	}
	mcDir := path.Dir(executable)

	cfgFile := path.Join(mcDir, ".crp-loader", "crp-loader-config.json")

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		fmt.Println("Ошибка чтения конфигурационного файла:", err)
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	err := write(*cfg)
	if err != nil {
		return err
	}
	return nil
}

func write(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	executable, err := os.Executable()
	if err != nil {
		fmt.Println("Ошибка определения пути к исполняемому файлу:", err)
		return err
	}

	mcDir := path.Dir(executable)

	configDir := path.Join(mcDir, ".crp-loader")

	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		return err
	}

	cfgFile := path.Join(configDir, "crp-loader-config.json")

	err = os.WriteFile(cfgFile, data, 0o644)
	if err != nil {
		return err
	}

	return nil
}
