package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Server struct {
	Address string `json:"address"`
}

type Config struct {
	ListenAddr          string    `json:"listen_addr"`
	Algorithm           string    `json:"algorithm"`
	HealthCheckInterval int       `json:"health_check_interval"`
	RequestTimeout      int       `json:"request_timeout"`
	Backends            []Server `json:"servers"`
}

// Load reads a JSON config file from the given path and returns a Config struct
func LoadConfigurationFile(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}

	// Set defaults if needed
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.Algorithm == "" {
		cfg.Algorithm = "round_robin"
	}
	if cfg.HealthCheckInterval <= 0 {
		cfg.HealthCheckInterval = 10
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 5
	}

	return &cfg, nil
}

// Returns the health check interval as a time.Duration
func (c *Config) HealthCheckDuration() time.Duration {
	return time.Duration(c.HealthCheckInterval) * time.Second
}

// Returns the request timeout as a time.Duration
func (c *Config) RequestTimeoutDuration() time.Duration {
	return time.Duration(c.RequestTimeout) * time.Second
}
