package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig 应用配置
type AppConfig struct {
	Nickname     string `json:"nickname"`
	DarkMode     bool   `json:"dark_mode"`
	JavaPath     string `json:"java_path,omitempty"`
	ServerMemoryMB int  `json:"server_memory_mb"`
	RelayServer  string `json:"relay_server,omitempty"`
	AllowRelay   bool   `json:"allow_community_relay"`
}

// Default 返回默认配置
func Default() *AppConfig {
	return &AppConfig{
		Nickname:       "MC玩家",
		DarkMode:       true,
		JavaPath:       "",
		ServerMemoryMB: 2048,
		RelayServer:    "",
		AllowRelay:     true,
	}
}

// Load 从文件加载配置
func Load() (*AppConfig, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), os.ErrNotExist
	}

	cfg := &AppConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return Default(), err
	}
	return cfg, nil
}

// Save 保存配置到文件
func (c *AppConfig) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)
}

func configDir() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".", err
		}
		return filepath.Join(home, ".mc-connector"), nil
	}
	return filepath.Join(appdata, "mc-connector"), nil
}
