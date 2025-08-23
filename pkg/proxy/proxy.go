package proxy

import (
	"io"
	"net"
	"time"

	"github.com/Gzimvra/golb/pkg/algorithms"
	"github.com/Gzimvra/golb/pkg/server"
	"github.com/Gzimvra/golb/pkg/utils"
)

// Proxy forwards connections to backend servers using RoundRobin
type Proxy struct {
	RoundRobin *algorithms.RoundRobin
	Timeout    time.Duration
}

// NewProxy creates a new Proxy instance
func NewProxy(pool *server.ServerPool, timeout time.Duration) *Proxy {
	return &Proxy{
		RoundRobin: algorithms.NewRoundRobin(pool),
		Timeout:    timeout,
	}
}

// Handle forwards the client connection to the selected backend
func (p *Proxy) Handle(clientConn net.Conn) {
	defer clientConn.Close()

	servers := p.RoundRobin.Pool.ListServers()
	if len(servers) == 0 {
		utils.Warn("No servers configured", nil)
		return
	}

	maxAttempts := len(servers) // try each server once
	var backendConn net.Conn
	var backend *server.Server
	var err error

	for range maxAttempts {
		backend = p.RoundRobin.Next()

		if backend == nil {
			// All servers marked dead, try on-demand probe
			for _, s := range servers {
				backendConn, err = net.DialTimeout("tcp", s.Address, p.Timeout)
				if err == nil {
					backend = s
					backend.MarkAlive()
					utils.Info("Recovered server via on-demand check", map[string]any{"server": s.Address})
					break
				}
			}
			if backendConn == nil {
				utils.Warn("No alive backend available", nil)
				return
			}
			break
		}

		backendConn, err = net.DialTimeout("tcp", backend.Address, p.Timeout)
		if err == nil {
			break
		}

		utils.Warn("Failed to connect to backend", map[string]any{"server": backend.Address, "error": err})
		backend.MarkDead() // mark dead so health checker will revalidate later
	}

	if backendConn == nil {
		utils.Warn("All backend connection attempts failed", nil)
		return
	}
	defer backendConn.Close()

	utils.Info("Forwarding request to backend", map[string]any{"server": backend.Address})

	// Proxy data: client <-> backend
	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn)
}

