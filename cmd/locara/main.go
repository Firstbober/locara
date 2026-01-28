package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Firstbober/locara/internal/config"
)

func main() {
	configPath := flag.String("config", config.DefaultConfigPath, "Path to configuration file")
	port := flag.Int("port", 0, "Server port (overrides config file)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to load config: %v", err)
	}

	if *port > 0 {
		if *port < 1 || *port > 65535 {
			log.Fatalf("[ERROR] Invalid port number: %d", *port)
		}
		cfg.Port = *port
	}

	log.Printf("[INFO] Starting Locara server on port %d", cfg.Port)
	log.Printf("[INFO] Using uploads directory: %s", cfg.UseDirectory)
	log.Printf("[INFO] Configured %d user(s)", len(cfg.Users))

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[INFO] Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Server failed: %v", err)
		}
	}()

	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[INFO] Received signal %v, shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server shutdown error: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] Server stopped gracefully")
}
