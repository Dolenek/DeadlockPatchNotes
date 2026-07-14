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

	forumURL := readStringEnv("PATCH_FORUM_URL", "https://forums.playdeadlock.com/forums/changelog.10/")
	steamNewsURL := readStringEnv("PATCH_STEAM_NEWS_URL", ingest.DefaultSteamNewsURL)

	maxPages := readIntEnv("PATCH_SYNC_MAX_PAGES", 20)
	timeoutSeconds := readIntEnv("PATCH_SYNC_TIMEOUT_SECONDS", 30)

	ctx := context.Background()
	database, err := db.OpenPostgres(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}
	stats, err := ingest.RunPatchSync(ctx, database, client, ingest.SyncConfig{
		ForumURL:     forumURL,
		SteamNewsURL: steamNewsURL,
		MaxPages:     maxPages,
	})
	if err != nil {
		log.Fatal(err)
	}

	printStats(stats)
}

func readStringEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func printStats(stats ingest.SyncStats) {
	fmt.Printf("sync complete: discovered=%d processed=%d failed=%d inserted=%d updated=%d\n",
		stats.DiscoveredThreads,
		stats.ProcessedThreads,
		stats.FailedThreads,
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
