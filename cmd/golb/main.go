package main

import (
	"fmt"
	"net"

	"github.com/Gzimvra/golb/pkg/config"
	"github.com/Gzimvra/golb/pkg/health"
	"github.com/Gzimvra/golb/pkg/server"
)

func main() {
    // Load the configuration file
	cfg, err := config.LoadConfigurationFile("./config.json")
	if err != nil {
		panic(err)
	}
	fmt.Println("Configuration File Successfully Loaded!")

	// Initialize server pool
	pool := &server.ServerPool{}
	for _, b := range cfg.Servers {
		pool.AddServer(&server.Server{
			Address: b.Address,
			Alive:   false, // explicitly mark as dead at startup
		})
	}

	// Start health checks
	hc := health.NewHealthChecker(pool, cfg.HealthCheckDuration(), cfg.RequestTimeoutDuration())

	// Run initial health check before starting ticker
	hc.CheckServers()
	fmt.Printf("Initial health check complete: %d/%d servers alive\n", pool.CountAlive(), len(pool.ListServers()))

	hc.Start()

	// Start a TCP listener
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Load balancer listening on %v\n\n", cfg.ListenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// spawn goroutine for each client
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	buf := make([]byte, 4096)
	n, _ := c.Read(buf)
	fmt.Println("Received:", string(buf[:n]))

	body := "Hello from the Load Balancer!"
	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n%s",
		len(body), body,
	)

	c.Write([]byte(response))
}
