package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"propagatorGo/internal/api/router"
	"propagatorGo/internal/config"
	"propagatorGo/internal/repository"
	"time"

	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	httpServer  *http.Server
	router      *mux.Router
	config      *config.Config
	articleRepo *repository.ExtArticleRepo
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, articleRepo *repository.ExtArticleRepo) *Server {
	r := router.Setup(cfg, articleRepo)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer:  srv,
		router:      r,
		config:      cfg,
		articleRepo: articleRepo,
	}
}

// Start begins serving API requests
func (s *Server) Start() error {
	log.Printf("API server starting on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("API server shutting down gracefully")
	return s.httpServer.Shutdown(ctx)
}

// Router returns the underlying router for testing or additional configuration
func (s *Server) Router() *mux.Router {
	return s.router
}

// Config returns the server's configuration
func (s *Server) Config() *config.Config {
	return s.config
}
