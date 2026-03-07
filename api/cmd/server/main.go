package main

import (
	"log"
	"net/http"
	"os"

	"deadlockpatchnotes/api/internal/httpapi"
	"deadlockpatchnotes/api/internal/patches"
)

func main() {
	addr := ":8080"
	if env := os.Getenv("API_ADDR"); env != "" {
		addr = env
	}

	store := patches.NewStore()
	handler := httpapi.NewRouter(store)

	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
