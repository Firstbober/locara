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
	"github.com/Firstbober/locara/internal/handlers"
	"github.com/Firstbober/locara/internal/templates"
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

	setupReverseProxy()

	tmpl, err := templates.ParseTemplatesFromFS()
	if err != nil {
		log.Fatalf("[ERROR] Failed to parse templates: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", loggingMiddleware(handlers.IndexHandler(tmpl, cfg)))
	mux.HandleFunc("GET /upload", loggingMiddleware(handlers.UploadHandler(tmpl, cfg)))
	mux.HandleFunc("GET /error", loggingMiddleware(handlers.ErrorHandler(tmpl, cfg)))
	mux.HandleFunc("POST /api/archive/create", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateArchiveHandler(w, r, cfg)
	}))
	mux.HandleFunc("GET /api/archives", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.ListArchivesHandler(w, r, cfg)
	}))
	mux.HandleFunc("GET /api/archive/{id}", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.DownloadArchiveHandler(w, r, cfg)
	}))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      loggingMiddlewareAll(mux),
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

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		clientIP := getClientIP(r)
		log.Printf("[INFO] %s %s from %s", r.Method, r.URL.Path, clientIP)

		next(w, r)

		log.Printf("[INFO] %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

func loggingMiddlewareAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		clientIP := getClientIP(r)
		log.Printf("[INFO] %s %s from %s", r.Method, r.URL.Path, clientIP)

		next.ServeHTTP(w, r)

		log.Printf("[INFO] %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
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

func setupReverseProxy() {
	if len(os.Getenv("TRUSTED_PROXIES")) > 0 {
		log.Printf("[INFO] Running behind reverse proxy, TRUSTED_PROXIES=%s", os.Getenv("TRUSTED_PROXIES"))
	}

	if os.Getenv("X_FORWARDED_FOR") != "" {
		log.Printf("[INFO] X-Forwarded-For detected, proxy mode enabled")
	}
}
