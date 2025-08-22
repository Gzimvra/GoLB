package server

import "sync"

// Server represents a backend server
type Server struct {
	Address string
	Alive   bool
	mu      sync.RWMutex
}

// ServerPool manages backend servers
type ServerPool struct {
	Servers []*Server
	mu      sync.RWMutex
}

// MarkAlive sets the server as alive
func (s *Server) MarkAlive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Alive = true
}

// MarkDead sets the server as dead
func (s *Server) MarkDead() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Alive = false
}

// IsAlive safely returns if the server is alive
func (s *Server) IsAlive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Alive
}

// AddServer adds a server to the pool
func (sp *ServerPool) AddServer(s *Server) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.Servers = append(sp.Servers, s)
}

// All returns all servers
func (sp *ServerPool) All() []*Server {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.Servers
}

