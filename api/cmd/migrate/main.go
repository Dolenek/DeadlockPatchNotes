package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"deadlockpatchnotes/api/internal/db"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
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
	if err := db.ConfigureRuntimeRoles(ctx, database, db.RuntimeRolePasswords{
		API:  os.Getenv("API_DB_PASSWORD"),
		Sync: os.Getenv("SYNC_DB_PASSWORD"),
	}); err != nil {
		return err
	}

	log.Print("database migrations and runtime roles are ready")
	return nil
}
