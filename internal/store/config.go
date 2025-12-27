package store

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Database file name
const DBFileName = "credentials.db"

// SMBConfig holds SMB sync configuration
type config struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Share    string `yaml:"share"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// Default configuration
var defaultConfig = config{
	Enabled:  false,
	Host:     "",
	Port:     "445",
	Share:    "",
	User:     "",
	Password: "",
}

var SMBConfig config

// LoadConfig loads configuration from the YAML file in .lockin directory
func LoadConfig() error {
	configPath := GetConfigPath()

	// If config doesn't exist, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		SMBConfig = defaultConfig
		return SaveConfig()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &SMBConfig); err != nil {
		return err
	}

	return nil
}

// SaveConfig saves the current configuration to the YAML file
func SaveConfig() error {
	configPath := GetConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(&SMBConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// GetConfigDir returns the path to the .lockin config directory
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".lockin"
	}
	return filepath.Join(homeDir, ".lockin")
}

// GetConfigPath returns the path to the config.yaml file
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.yaml")
}

// GetDBPath returns the path to the credentials database
func GetDBPath() string {
	return filepath.Join(GetConfigDir(), DBFileName)
}

// IsSMBEnabled returns true if SMB sync is enabled in config
func IsSMBEnabled() bool {
	return SMBConfig.Enabled
}
