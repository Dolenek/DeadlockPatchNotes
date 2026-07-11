package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"deadlockpatchnotes/api/internal/db"
	"deadlockpatchnotes/api/internal/httpapi"
	"deadlockpatchnotes/api/internal/patches"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	addr := ":8080"
	if env := os.Getenv("API_ADDR"); env != "" {
		addr = env
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	database, err := db.OpenPostgres(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.ApplyMigrations(ctx, database); err != nil {
		return err
	}

	cacheTTL := readDurationEnv("API_READ_CACHE_TTL", 10*time.Minute)
	store := patches.NewPostgresStore(database, cacheTTL)
	handler := httpapi.NewRouter(store, database.PingContext)
	server := newHTTPServer(addr, handler)
	log.Printf("api listening on %s", addr)
	return serveUntilCanceled(ctx, server)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

func serveUntilCanceled(ctx context.Context, server *http.Server) error {
	serveCtx, cancelServe := context.WithCancel(ctx)
	defer cancelServe()
	shutdownComplete := make(chan struct{})
	go func() {
		defer close(shutdownComplete)
		<-serveCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	cancelServe()
	<-shutdownComplete
	return nil
}

func readDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}
