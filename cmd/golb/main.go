package main

import (
	"net"

	"github.com/Gzimvra/golb/pkg/config"
	"github.com/Gzimvra/golb/pkg/health"
	"github.com/Gzimvra/golb/pkg/middleware/ipfilter"
	"github.com/Gzimvra/golb/pkg/middleware/ratelimiter"
	"github.com/Gzimvra/golb/pkg/proxy"
	"github.com/Gzimvra/golb/pkg/server"
	"github.com/Gzimvra/golb/pkg/utils/logger"
	"github.com/Gzimvra/golb/pkg/utils/netutils"
)

func main() {
	// Initialize the logger
	logger.Init()
	logger.Info("Starting GoLB load balancer", nil)

	// Load configuration
	cfg, err := config.LoadConfigurationFile("./config.json")
	if err != nil {
		logger.Error("Failed to load configuration", nil)
		panic(err)
	}
	logger.Info("Configuration successfully loaded", nil)

	// Initialize server pool
	pool := &server.ServerPool{}
	for _, b := range cfg.Servers {
		pool.AddServer(&server.Server{
			Address: b.Address,
			Alive:   false, // mark dead at startup
		})
	}

	// Start health checks
	hc := health.NewHealthChecker(pool, cfg.HealthCheckDuration(), cfg.RequestTimeoutDuration())
	hc.CheckServers() // initial check
	hc.StartHealthChecker()
	logger.Info("Health checker started", nil)

	// Initialize proxy
	prx := proxy.NewProxy(pool, cfg.RequestTimeoutDuration())

	// Initialize rate limiter
	rl := ratelimiter.NewRateLimiter(cfg.MaxConcurrentConnections, cfg.MaxConnectionsPerMinute)

	// Initialize IP filter
	ipf := ipfilter.NewIPFilter(cfg)

	// Start TCP listener
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		logger.Error("Failed to start TCP listener", nil)
		panic(err)
	}
	defer listener.Close()
	logger.Info("Load balancer listening on "+cfg.ListenAddr, nil)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("Error accepting connection", nil)
			continue
		}

		clientIP := netutils.GetClientIP(conn)

		handleConnection(conn, clientIP, ipf, rl, prx)
	}
}

func handleConnection(conn net.Conn, clientIP string, ipf *ipfilter.IPFilter, rl *ratelimiter.RateLimiter, prx *proxy.Proxy) {
	// IP filter
	if !ipf.Allow(clientIP) {
		logger.Warn("Connection rejected by IP filter", map[string]any{"ip": clientIP})
		conn.Close()
		return
	}

	// Rate limiter
	if !rl.Allow(clientIP) {
		logger.Warn("Connection rejected due to rate limiting", map[string]any{"ip": clientIP})
		conn.Close()
		return
	}

	// Handle connection
	go func(c net.Conn, ip string) {
		defer rl.Done(ip)
		prx.Handle(c)
	}(conn, clientIP)
}
