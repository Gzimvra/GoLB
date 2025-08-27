package ipfilter

import (
	"strings"

	"github.com/Gzimvra/golb/pkg/config"
	"github.com/Gzimvra/golb/pkg/utils/logger"
	"slices"
)

// IPFilter handles allow/deny rules for incoming connections
type IPFilter struct {
	mode string   // "allow", "deny", "none"
	list []string // list of IPs
}

// NewIPFilter creates a new IPFilter from config
func NewIPFilter(cfg *config.Config) *IPFilter {
	return &IPFilter{
		mode: strings.ToLower(cfg.IPFilterMode),
		list: cfg.IPFilterList,
	}
}

// Allow checks if a client IP can connect
func (f *IPFilter) Allow(ip string) bool {
	ip = strings.TrimSpace(ip)

	// If filtering is disabled, allow all
	if f.mode == "none" || len(f.list) == 0 {
		return true
	}

	switch f.mode {
	case "allow":
		if slices.Contains(f.list, ip) {
			return true
		}
		logger.Warn("Connection rejected by allowlist", map[string]any{"ip": ip})
		return false
	case "deny":
		if slices.Contains(f.list, ip) {
			logger.Warn("Connection rejected by denylist", map[string]any{"ip": ip})
			return false
		}
		return true
	default:
		// Unknown mode, be safe and allow
		logger.Warn("Detected unknown ip-filter mode, defaulting to allow-all", nil)
		return true
	}
}
