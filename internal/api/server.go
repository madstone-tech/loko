package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/madstone-tech/loko/internal/api/handlers"
	"github.com/madstone-tech/loko/internal/api/middleware"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ServerConfig holds configuration for the API server.
type ServerConfig struct {
	Port        int
	ProjectRoot string
	APIKey      string // Optional API key for authentication
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

// DefaultConfig returns a default server configuration.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Port:         8081,
		ProjectRoot:  ".",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
}

// Server is the HTTP API server for loko.
type Server struct {
	config     ServerConfig
	repo       usecases.ProjectRepository
	httpServer *http.Server
	startTime  time.Time
}

// NewServer creates a new API server.
func NewServer(config ServerConfig, repo usecases.ProjectRepository) *Server {
	return &Server{
		config:    config,
		repo:      repo,
		startTime: time.Now(),
	}
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Create handlers
	h := handlers.NewHandlers(s.config.ProjectRoot, s.repo)

	// Health check (no auth required)
	mux.HandleFunc("GET /health", s.handleHealth)

	// API v1 routes
	mux.HandleFunc("GET /api/v1/project", h.GetProject)
	mux.HandleFunc("GET /api/v1/systems", h.ListSystems)
	mux.HandleFunc("GET /api/v1/systems/{id}", h.GetSystem)
	mux.HandleFunc("POST /api/v1/build", h.TriggerBuild)
	mux.HandleFunc("GET /api/v1/build/{id}", h.GetBuildStatus)
	mux.HandleFunc("GET /api/v1/validate", h.Validate)

	// Apply middleware chain
	var handler http.Handler = mux

	// Add auth middleware if API key is configured
	if s.config.APIKey != "" {
		handler = middleware.Auth(s.config.APIKey)(handler)
	}

	// Add common middleware
	handler = middleware.Logger(handler)
	handler = middleware.CORS(handler)
	handler = middleware.Recovery(handler)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Start server
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return s.Shutdown()
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// handleHealth handles GET /health.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime).Round(time.Second).String()

	resp := handlers.HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
		Uptime:  uptime,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
