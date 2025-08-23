package algorithms

import (
	"github.com/Gzimvra/golb/pkg/server"
	"sync"
)

// RoundRobin implements a round-robin load balancing algorithm
type RoundRobin struct {
	Pool    *server.ServerPool
	current int
	mu      sync.Mutex
}

// Next returns the next alive server using round-robin
func (rr *RoundRobin) Next() *server.Server {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	servers := rr.Pool.ListServers()
	if len(servers) == 0 {
		return nil
	}

	start := rr.current
	for {
		s := servers[rr.current]
		rr.current = (rr.current + 1) % len(servers)

		if s.IsAlive() {
			return s
		}

		if rr.current == start {
			return nil // no alive servers
		}
	}
}

// NewRoundRobin creates a new RoundRobin instance
func NewRoundRobin(pool *server.ServerPool) *RoundRobin {
	return &RoundRobin{
		Pool: pool,
	}
}
