package main

import (
	"fmt"
	"net"

	"github.com/Gzimvra/golb/pkg/config"
	"github.com/Gzimvra/golb/pkg/health"
	"github.com/Gzimvra/golb/pkg/proxy"
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
	hc.CheckServers() // Run initial health check before starting ticker
	hc.StartHealthChecker()

	// Initialize proxy
	prx := proxy.NewProxy(pool, cfg.RequestTimeoutDuration())

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

		// Forward connection through the proxy
		go prx.Handle(conn)
	}
}
