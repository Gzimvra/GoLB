package health

import (
	"net"
	"time"

	"github.com/Gzimvra/golb/pkg/server"
	"github.com/Gzimvra/golb/pkg/utils/logger"
)

// HealthChecker periodically checks the health of backend servers
type HealthChecker struct {
	Pool     *server.ServerPool
	Interval time.Duration
	Timeout  time.Duration
}

// NewHealthChecker creates a new HealthChecker
func NewHealthChecker(pool *server.ServerPool, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		Pool:     pool,
		Interval: interval,
		Timeout:  timeout,
	}
}

// StartHealthChecker launches the health checks in a separate goroutine
func (hc *HealthChecker) StartHealthChecker() {
	go func() {
		ticker := time.NewTicker(hc.Interval)
		defer ticker.Stop()

		for range ticker.C {
			hc.CheckServers()
		}
	}()
}

// CheckServers tests each server and marks it alive or dead
func (hc *HealthChecker) CheckServers() {
	logger.Info("Health check starting", nil)
	for _, s := range hc.Pool.ListServers() {
		alive := hc.isAlive(s.Address)
		if alive {
			s.MarkAlive()
		} else {
			s.MarkDead()
		}
		logger.Info("Health check", map[string]any{
			"server": s.Address,
			"alive":  alive,
		})
	}

	logger.Info("Health check completed", map[string]any{
		"alive_count": hc.Pool.CountAlive(),
		"total":       len(hc.Pool.ListServers()),
	})
}

// isAlive tries to open a TCP connection to the server
func (hc *HealthChecker) isAlive(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, hc.Timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
