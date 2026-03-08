package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"deadlockpatchnotes/api/internal/db"
	"deadlockpatchnotes/api/internal/ingest"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	forumURL := os.Getenv("PATCH_FORUM_URL")
	if forumURL == "" {
		forumURL = "https://forums.playdeadlock.com/forums/changelog.10/"
	}

	maxPages := readIntEnv("PATCH_SYNC_MAX_PAGES", 20)
	timeoutSeconds := readIntEnv("PATCH_SYNC_TIMEOUT_SECONDS", 30)

	ctx := context.Background()
	database, err := db.OpenPostgres(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(ctx, database); err != nil {
		log.Fatal(err)
	}

	client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}
	stats, err := ingest.RunPatchSync(ctx, database, client, ingest.SyncConfig{
		ForumURL: forumURL,
		MaxPages: maxPages,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sync complete: discovered=%d processed=%d inserted=%d updated=%d\n",
		stats.DiscoveredThreads,
		stats.ProcessedThreads,
		stats.InsertedPatches,
		stats.UpdatedPatches,
	)
}

func readIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
