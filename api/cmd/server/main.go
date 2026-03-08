package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	store := patches.NewPostgresStore(database)
	handler := httpapi.NewRouter(store)

	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
