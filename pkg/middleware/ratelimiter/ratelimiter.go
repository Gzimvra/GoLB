package ratelimiter

import (
	"sync"
	"time"
)

// ClientStats tracks connection info for a single client IP
type ClientStats struct {
	Concurrent int
	Timestamps []time.Time
}

// RateLimiter manages per-IP connection limits
type RateLimiter struct {
	sync.Mutex
	Clients                 map[string]*ClientStats
	MaxConcurrent           int
	MaxConnectionsPerMinute int
	Window                  time.Duration
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(maxConc, maxPerMin int) *RateLimiter {
	return &RateLimiter{
		Clients:                 make(map[string]*ClientStats),
		MaxConcurrent:           maxConc,
		MaxConnectionsPerMinute: maxPerMin,
		Window:                  time.Minute,
	}
}

// Allow checks if a new connection is allowed and updates counters if yes
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	stats, exists := rl.Clients[clientIP]
	if !exists {
		stats = &ClientStats{}
		rl.Clients[clientIP] = stats
	}

	// Remove timestamps outside the sliding window
	validTimestamps := []time.Time{}
	for _, t := range stats.Timestamps {
		if now.Sub(t) <= rl.Window {
			validTimestamps = append(validTimestamps, t)
		}
	}
	stats.Timestamps = validTimestamps

	// Check concurrent connections
	if stats.Concurrent >= rl.MaxConcurrent {
		return false
	}

	// Check rate (connections per minute)
	if len(stats.Timestamps) >= rl.MaxConnectionsPerMinute {
		return false
	}

	// Allow connection
	stats.Concurrent++
	stats.Timestamps = append(stats.Timestamps, now)
	return true
}

// Done should be called when a connection closes
func (rl *RateLimiter) Done(clientIP string) {
	rl.Lock()
	defer rl.Unlock()

	if stats, exists := rl.Clients[clientIP]; exists && stats.Concurrent > 0 {
		stats.Concurrent--
	}
}
