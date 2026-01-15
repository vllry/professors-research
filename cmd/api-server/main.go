package main

import (
	"log"
	"os"

	"github.com/vllry/professors-research/internal/api-server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := apiserver.Config{
		Port:    port,
		DataDir: "data",
	}

	server, err := apiserver.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
