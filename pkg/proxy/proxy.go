package proxy

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Gzimvra/golb/pkg/algorithms"
	"github.com/Gzimvra/golb/pkg/server"
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
	backend := p.RoundRobin.Next()
	if backend == nil {
		fmt.Println("No alive backend available")
		clientConn.Close()
		return
	}

	// Connect to backend
	backendConn, err := net.DialTimeout("tcp", backend.Address, p.Timeout)
	if err != nil {
		fmt.Printf("Failed to connect to backend %s: %v\n", backend.Address, err)
		clientConn.Close()
		return
	}
	defer backendConn.Close()
	defer clientConn.Close()

	fmt.Printf("Forwarding request to backend %s\n", backend.Address)

	// Start bidirectional copy
	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn) // block until done
}

