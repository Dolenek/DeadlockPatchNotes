package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"deadlockpatchnotes/api/internal/patches"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	store *patches.Store
}

func NewRouter(store *patches.Store) http.Handler {
	api := &API{store: store}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/api/healthz", api.healthz)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/patches", api.listPatches)
		v1.Get("/patches/{slug}", api.getPatch)
	})

	return r
}

func (a *API) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) listPatches(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 12)

	if limit > 50 {
		limit = 50
	}

	payload := a.store.List(page, limit)
	writeJSON(w, http.StatusOK, payload)
}

func (a *API) getPatch(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing slug")
		return
	}

	patch, err := a.store.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, patches.ErrPatchNotFound) {
			writeError(w, http.StatusNotFound, "patch not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load patch")
		return
	}

	writeJSON(w, http.StatusOK, patch)
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
