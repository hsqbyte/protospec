// Package cloud provides PSL Cloud API server for protocol-as-a-service.
package cloud

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Server is the PSL Cloud API server.
type Server struct {
	lib       *protocol.Library
	addr      string
	apiKeys   map[string]*Tenant
	mu        sync.RWMutex
	rateLimit map[string]*rateLimiter
}

// Tenant represents a cloud tenant.
type Tenant struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Quota int    `json:"quota"` // requests per minute
}

type rateLimiter struct {
	count   int
	resetAt time.Time
	limit   int
}

// NewServer creates a new cloud API server.
func NewServer(lib *protocol.Library, addr string) *Server {
	return &Server{
		lib:       lib,
		addr:      addr,
		apiKeys:   make(map[string]*Tenant),
		rateLimit: make(map[string]*rateLimiter),
	}
}

// AddTenant adds a tenant with an API key.
func (s *Server) AddTenant(apiKey string, tenant *Tenant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apiKeys[apiKey] = tenant
}

// Start starts the cloud API server.
func (s *Server) Start() error {
	// Add default demo tenant
	s.AddTenant("demo-key", &Tenant{ID: "demo", Name: "Demo", Quota: 100})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/decode/", s.handleDecode)
	mux.HandleFunc("/api/v1/encode/", s.handleEncode)
	mux.HandleFunc("/api/v1/protocols", s.handleListProtocols)
	mux.HandleFunc("/api/v1/health", s.handleHealth)

	fmt.Printf("PSL Cloud API running at http://%s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	protoName := strings.TrimPrefix(r.URL.Path, "/api/v1/decode/")
	if protoName == "" {
		writeError(w, http.StatusBadRequest, "protocol name required")
		return
	}

	var req struct {
		Hex string `json:"hex"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	data, err := hex.DecodeString(strings.ReplaceAll(req.Hex, " ", ""))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hex")
		return
	}

	result, err := s.lib.Decode(protoName, data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, map[string]any{"success": true, "data": result.Packet, "bytes_read": result.BytesRead})
}

func (s *Server) handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	protoName := strings.TrimPrefix(r.URL.Path, "/api/v1/encode/")
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	encoded, err := s.lib.Encode(protoName, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, map[string]any{"success": true, "hex": hex.EncodeToString(encoded), "length": len(encoded)})
}

func (s *Server) handleListProtocols(w http.ResponseWriter, r *http.Request) {
	names := s.lib.AllNames()
	writeJSON(w, map[string]any{"success": true, "protocols": names, "count": len(names)})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"status": "ok", "version": "0.50.0"})
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"success": false, "error": msg})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
