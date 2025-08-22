package health

import (
	"fmt"
	"net"
	"time"

	"github.com/Gzimvra/golb/pkg/server"
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

// Start launches the health checks in a separate goroutine
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
	for _, s := range hc.Pool.ListServers() {
		alive := hc.isAlive(s.Address)
		if alive {
			s.MarkAlive()
		} else {
			s.MarkDead()
		}
		fmt.Printf("Health check: %s is alive=%v\n", s.Address, alive)
	}
	fmt.Printf("Health check: Completed successfully with %d/%d servers alive\n",
		hc.Pool.CountAlive(),
		len(hc.Pool.ListServers()),
	)
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
