package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/Gzimvra/golb/pkg/utils"
)

type Server struct {
	Address string `json:"address"`
}

type Config struct {
	ListenAddr               string   `json:"listen_addr"`
	Algorithm                string   `json:"algorithm"`
	HealthCheckInterval      int      `json:"health_check_interval"`
	RequestTimeout           int      `json:"request_timeout"`
	MaxConcurrentConnections int      `json:"max_concurrent_connections"`
	MaxConnectionsPerMinute  int      `json:"max_connections_per_minute"`
	Servers                  []Server `json:"servers"`
}

// LoadConfigurationFile reads a JSON config file from the given path and returns a Config struct
func LoadConfigurationFile(path string) (*Config, error) {
	log := utils.GetLogger()

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error("Cannot read config file", nil)
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		log.Error("Invalid config format", nil)
		return nil, err
	}

	// Set defaults if needed
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
		log.Warn("ListenAddr not set in config, using default :8080", nil)
	}
	if cfg.Algorithm == "" {
		cfg.Algorithm = "round_robin"
		log.Warn("Algorithm not set in config, using default round_robin", nil)
	}
	if cfg.HealthCheckInterval <= 0 {
		cfg.HealthCheckInterval = 10
		log.Warn("HealthCheckInterval invalid or not set, using default 10s", nil)
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 5
		log.Warn("RequestTimeout invalid or not set, using default 5s", nil)
	}
	if cfg.MaxConcurrentConnections <= 0 {
		cfg.MaxConcurrentConnections = 10
		log.Warn("MaxConcurrentConnections invalid or not set, using default 10", nil)
	}
	if cfg.MaxConnectionsPerMinute <= 0 {
		cfg.MaxConnectionsPerMinute = 50
		log.Warn("MaxConnectionsPerMinute invalid or not set, using default 50", nil)
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

