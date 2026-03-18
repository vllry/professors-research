package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vllry/professors-research/internal/detailedcardcache"
	"github.com/vllry/professors-research/internal/matchups"
)

// Server represents the HTTP server
type Server struct {
	httpServer    *http.Server
	detailedCards *detailedcardcache.DetailedCardCache
	archetypes    []matchups.Archetype
	tournaments   map[string]tournamentInfo
	dataDir       string
}

// Config holds server configuration
type Config struct {
	Port    string
	DataDir string // Path to the data directory (defaults to "data")
}

// NewServer creates a new Server instance
func NewServer(cfg Config) (*Server, error) {
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	dataDir := cfg.DataDir
	if dataDir == "" {
		dataDir = "data"
	}

	// Start loading DetailedCard cache asynchronously
	cache := detailedcardcache.NewDetailedCardCache(dataDir)

	archetypes, err := matchups.LoadArchetypes(filepath.Join(dataDir, "archetypes.json"))
	if err != nil {
		log.Printf("Warning: failed to load archetypes: %v", err)
	}

	tournaments := loadTournamentRegistry(dataDir)

	mux := http.NewServeMux()

	server := &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		detailedCards: cache,
		archetypes:    archetypes,
		tournaments:   tournaments,
		dataDir:       dataDir,
	}

	// Health check endpoints - support multiple paths for compatibility
	// GKE Gateway NEG health checks may use /, /health, or /api/health
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodGet {
			// Root path health check for NEG
			server.handleHealth(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/prize-odds", server.handlePrizeOdds)
	mux.HandleFunc("/api/start-odds", server.handleStartOdds)
	mux.HandleFunc("/api/draw-supporter-odds", server.handleDrawSupporterOdds)
	mux.HandleFunc("/api/matchup-stats", server.handleMatchupStats)
	mux.HandleFunc("/api/deck-variants", server.handleDeckVariants)
	mux.HandleFunc("/api/tournaments", server.handleTournaments)
	mux.HandleFunc("/api/archetypes", server.handleArchetypes)

	// Apply security headers middleware
	server.httpServer.Handler = addSecurityHeaders(mux)

	return server, nil
}

// Run starts the server and blocks until shutdown
func (s *Server) Run() error {
	log.Printf("Server starting on port %s", s.httpServer.Addr[1:])

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}
