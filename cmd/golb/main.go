package main

import (
	"net"

	"github.com/Gzimvra/golb/pkg/config"
	"github.com/Gzimvra/golb/pkg/health"
	"github.com/Gzimvra/golb/pkg/middleware"
	"github.com/Gzimvra/golb/pkg/proxy"
	"github.com/Gzimvra/golb/pkg/server"
	"github.com/Gzimvra/golb/pkg/utils"
)

func main() {
	// Initialize the logger
	utils.Init()

	// Load the configuration file
	cfg, err := config.LoadConfigurationFile("./config.json")
	if err != nil {
		utils.Error("Failed to load configuration", nil)
		panic(err)
	}
	utils.Info("Configuration file successfully loaded", nil)

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
	utils.Info("Health checker started", nil)

	// Initialize proxy
	prx := proxy.NewProxy(pool, cfg.RequestTimeoutDuration())

	// Initialize rate limiter middleware
	rateLimiter := middleware.NewRateLimiter(cfg.MaxConcurrentConnections, cfg.MaxConnectionsPerMinute)

	// Start TCP listener
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		utils.Error("Failed to start TCP listener", nil)
		panic(err)
	}
	defer listener.Close()
	utils.Info("Load balancer listening on "+cfg.ListenAddr, nil)

	for {
		conn, err := listener.Accept()
		if err != nil {
			utils.Warn("Error accepting connection", nil)
			continue
		}

		clientIP := middleware.GetClientIP(conn)

		// Rate limiting check
		if !rateLimiter.Allow(clientIP) {
			utils.Warn("Connection rejected due to rate limiting", map[string]any{"ip": clientIP})
			conn.Close()
			continue
		}

		// Make sure Done is called when the connection closes
		go func(c net.Conn, ip string) {
			defer rateLimiter.Done(ip)
			prx.Handle(c)
		}(conn, clientIP)
	}
}

