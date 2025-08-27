package netutils

import "net"

// GetClientIP extracts the remote IP from net.Conn
func GetClientIP(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	// Split host:port
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr // fallback to full addr
	}
	return host
}
