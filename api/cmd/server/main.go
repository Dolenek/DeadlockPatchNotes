package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"deadlockpatchnotes/api/internal/db"
	"deadlockpatchnotes/api/internal/httpapi"
	"deadlockpatchnotes/api/internal/patches"
)

func main() {
	addr := ":8080"
	if env := os.Getenv("API_ADDR"); env != "" {
		addr = env
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	database, err := db.OpenPostgres(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(ctx, database); err != nil {
		log.Fatal(err)
	}

	cacheTTL := readDurationEnv("API_READ_CACHE_TTL", 10*time.Minute)
	store := patches.NewPostgresStore(database, cacheTTL)
	handler := httpapi.NewRouter(store)

	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
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
