package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Configuration represents the application configuration
type Configuration struct {
	Server struct {
		Port int `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Projects struct {
		Directory string `yaml:"directory"`
	} `yaml:"projects"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Balancer struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"balancer"`
	Scanning struct {
		AutoScanOnAdd bool `yaml:"auto_scan_on_add"`
	} `yaml:"scanning"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Configuration {
	cfg := &Configuration{}
	
	// Set default values
	cfg.Server.Port = 8080
	cfg.Server.Host = "0.0.0.0"
	cfg.Projects.Directory = "./projects"
	cfg.Database.Path = "./syncx.db"
	cfg.Balancer.Endpoints = []string{"http://localhost:8081", "http://localhost:8082"}
	cfg.Scanning.AutoScanOnAdd = true
	
	return cfg
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Configuration, error) {
	// Start with default configuration
	config := DefaultConfig()
	
	// If config file exists, load it
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		
		err = yaml.Unmarshal(data, config)
		if err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
		
		log.Printf("Loaded configuration from %s", path)
	} else {
		// Config file doesn't exist, create one with defaults
		err := SaveConfig(config, path)
		if err != nil {
			log.Printf("Warning: Could not save default config to %s: %v", path, err)
		} else {
			log.Printf("Created default configuration at %s", path)
		}
	}
	
	return config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Configuration, path string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
} 