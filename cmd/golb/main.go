package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	// Initialize logger
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
			Alive:   false,
		})
	}

	// Start health checks
	hc := health.NewHealthChecker(pool, cfg.HealthCheckDuration(), cfg.RequestTimeoutDuration())
	hc.CheckServers()
	hc.StartHealthChecker()
	logger.Info("Health checker started", nil)

	// Initialize proxy
	prx := proxy.NewProxy(pool, cfg.RequestTimeoutDuration())

	// Initialize rate limiter and IP filter
	rl := ratelimiter.NewRateLimiter(cfg.MaxConcurrentConnections, cfg.MaxConnectionsPerMinute)
	ipf := ipfilter.NewIPFilter(cfg)

	// TCP listener
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		logger.Error("Failed to start TCP listener", nil)
		panic(err)
	}
	defer listener.Close()
	logger.Info("Load balancer listening on "+cfg.ListenAddr, nil)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received", nil)
		cancel()
		listener.Close() // stop accepting new connections
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				// Shutdown in progress
				wg.Wait()
				logger.Info("All connections finished. Exiting.", nil)
				return
			default:
				logger.Warn("Error accepting connection", nil)
				continue
			}
		}

		clientIP := netutils.GetClientIP(conn)

		wg.Add(1)
		go func(c net.Conn, ip string) {
			defer wg.Done()
			handleConnection(c, ip, ipf, rl, prx)
		}(conn, clientIP)
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

	// Handle connection synchronously
	defer rl.Done(clientIP)
	prx.Handle(conn)
}

