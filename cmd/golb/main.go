package main

import (
	"context"
	"crypto/tls"
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

var activeConns sync.Map // Track all active connections

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

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health checks with context
	hc := health.NewHealthChecker(pool, cfg.HealthCheckDuration(), cfg.RequestTimeoutDuration())
	hc.CheckServers()
	hc.StartHealthChecker(ctx)
	logger.Info("Health checker started", nil)

	// Initialize proxy, rate limiter, and IP filter
	prx := proxy.NewProxy(pool, cfg.RequestTimeoutDuration())
	rl := ratelimiter.NewRateLimiter(cfg.MaxConcurrentConnections, cfg.MaxConnectionsPerMinute)
	ipf := ipfilter.NewIPFilter(cfg)

	// Start listener (TCP or TLS)
	listener, err := startListener(cfg)
	if err != nil {
		logger.Error("Failed to start listener", map[string]any{"err": err.Error()})
		panic(err)
	}
	defer listener.Close()

	var wg sync.WaitGroup

	// Listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received", nil)
		cancel()
		listener.Close() // stop accepting new connections

		// Close all active connections
		activeConns.Range(func(key, _ any) bool {
			conn := key.(net.Conn)
			conn.Close()
			return true
		})
	}()

	// Accept loop
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
				logger.Warn("Error accepting connection", map[string]any{"err": err.Error()})
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
	// Track active connection
	activeConns.Store(conn, struct{}{})
	defer activeConns.Delete(conn)
	defer conn.Close()
	defer rl.Done(clientIP)

	// IP filter
	if !ipf.Allow(clientIP) {
		logger.Warn("Connection rejected by IP filter", map[string]any{"ip": clientIP})
		return
	}

	// Rate limiter
	if !rl.Allow(clientIP) {
		logger.Warn("Connection rejected due to rate limiting", map[string]any{"ip": clientIP})
		return
	}

	// Handle the request
	prx.Handle(conn)
}

// startListener chooses between plain TCP and TLS
func startListener(cfg *config.Config) (net.Listener, error) {
	if cfg.AcceptTLS {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			logger.Error("Failed to load TLS certificate or key", map[string]any{
				"cert": cfg.TLSCertFile,
				"key":  cfg.TLSKeyFile,
				"err":  err.Error(),
			})
			return nil, err
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		logger.Info("TLS enabled. Listening on "+cfg.ListenAddr, nil)
		return tls.Listen("tcp", cfg.ListenAddr, tlsConfig)
	}

	logger.Info("Plain TCP mode. Listening on "+cfg.ListenAddr, nil)
	return net.Listen("tcp", cfg.ListenAddr)
}

