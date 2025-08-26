package proxy

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/Gzimvra/golb/pkg/algorithms"
	"github.com/Gzimvra/golb/pkg/server"
	"github.com/Gzimvra/golb/pkg/utils/logger"
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

// Handle is the entry point for a client connection
func (p *Proxy) Handle(clientConn net.Conn) {
	defer clientConn.Close()

	servers := p.RoundRobin.Pool.ListServers()
	if len(servers) == 0 {
		logger.Warn("No servers configured", nil)
		return
	}

	backendConn, backend := p.getBackendConnection(servers)
	if backendConn == nil {
		logger.Warn("All backend connection attempts failed", nil)
		return
	}
	defer backendConn.Close()

	logger.Info("Forwarding request to backend", map[string]any{"server": backend.Address})
	p.pipeTraffic(clientConn, backendConn)
}

// getBackendConnection selects and connects to a backend server
func (p *Proxy) getBackendConnection(servers []*server.Server) (net.Conn, *server.Server) {
	maxAttempts := len(servers)
	var backendConn net.Conn
	var backend *server.Server
	var err error

	for range maxAttempts {
		backend = p.RoundRobin.Next()
		if backend == nil {
			// All servers marked dead, try recovery
			return p.recoverBackend(servers)
		}

		backendConn, err = net.DialTimeout("tcp", backend.Address, p.Timeout)
		if err == nil {
			return backendConn, backend
		}

		logger.Warn("Failed to connect to backend", map[string]any{
			"server": backend.Address,
			"error":  err,
		})
		backend.MarkDead()
	}

	return nil, nil
}

// recoverBackend tries to reconnect to servers previously marked as dead
func (p *Proxy) recoverBackend(servers []*server.Server) (net.Conn, *server.Server) {
	for _, s := range servers {
		backendConn, err := net.DialTimeout("tcp", s.Address, p.Timeout)
		if err == nil {
			s.MarkAlive()
			logger.Info("Recovered server via on-demand check", map[string]any{"server": s.Address})
			return backendConn, s
		}
	}
	logger.Warn("No alive backend available", nil)
	return nil, nil
}

// pipeTraffic proxies data between client and backend safely
func (p *Proxy) pipeTraffic(clientConn, backendConn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	// client -> backend
	go func() {
		defer wg.Done()
		io.Copy(backendConn, clientConn)
		if tcpConn, ok := backendConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite() // signal EOF to backend
		}
	}()

	// backend -> client
	go func() {
		defer wg.Done()
		io.Copy(clientConn, backendConn)
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite() // signal EOF to client
		}
	}()

	wg.Wait() // wait for both directions to finish
}
