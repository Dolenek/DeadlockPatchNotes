package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"deadlockpatchnotes/api/internal/patches"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	store patches.Repository
}

func NewRouter(store patches.Repository) http.Handler {
	api := &API{store: store}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/api/healthz", api.healthz)
	r.Get("/api/scalar", api.scalarDocs)
	r.Get("/api/openapi.json", api.openapiSpec)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/days-since-last-update", api.daysSinceLastUpdate)
		v1.Get("/patches", api.listPatches)
		v1.Get("/patches/rss.xml", api.patchFeedRSS)
		v1.Get("/patches/{slug}", api.getPatch)
		v1.Get("/heroes", api.listHeroes)
		v1.Get("/heroes/{heroSlug}/changes", api.getHeroChanges)
		v1.Get("/heroes/{heroSlug}/rss.xml", api.heroFeedRSS)
		v1.Get("/heroes/{heroSlug}/days-without-update/rss.xml", api.heroDaysWithoutUpdateRSS)
		v1.Get("/items", api.listItems)
		v1.Get("/items/{itemSlug}/changes", api.getItemChanges)
		v1.Get("/spells", api.listSpells)
		v1.Get("/spells/{spellSlug}/changes", api.getSpellChanges)
	})

	return r
}

func (a *API) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) scalarDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(scalarHTML)
}

func (a *API) openapiSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpecJSON)
}

func (a *API) listPatches(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 12)

	if limit > 50 {
		limit = 50
	}

	payload, err := a.store.List(page, limit)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list patches")
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (a *API) getPatch(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		writeError(w, r, http.StatusBadRequest, "missing_path_param", "missing slug")
		return
	}

	patch, err := a.store.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, patches.ErrPatchNotFound) {
			writeError(w, r, http.StatusNotFound, "resource_not_found", "patch not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load patch")
		return
	}

	writeJSON(w, http.StatusOK, patch)
}

func (a *API) listHeroes(w http.ResponseWriter, r *http.Request) {
	payload, err := a.store.ListHeroes()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list heroes")
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *API) getHeroChanges(w http.ResponseWriter, r *http.Request) {
	heroSlug := strings.TrimSpace(chi.URLParam(r, "heroSlug"))
	if heroSlug == "" {
		writeError(w, r, http.StatusBadRequest, "missing_path_param", "missing hero slug")
		return
	}

	from, err := parseTimeQuery(r, "from", true)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid from query value")
		return
	}
	to, err := parseTimeQuery(r, "to", false)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid to query value")
		return
	}

	payload, err := a.store.GetHeroChanges(patches.HeroChangesQuery{
		HeroSlug: heroSlug,
		Skill:    strings.TrimSpace(r.URL.Query().Get("skill")),
		From:     from,
		To:       to,
	})
	if err != nil {
		if errors.Is(err, patches.ErrHeroNotFound) {
			writeError(w, r, http.StatusNotFound, "resource_not_found", "hero not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load hero changes")
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (a *API) listItems(w http.ResponseWriter, r *http.Request) {
	payload, err := a.store.ListItems()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list items")
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *API) getItemChanges(w http.ResponseWriter, r *http.Request) {
	itemSlug := strings.TrimSpace(chi.URLParam(r, "itemSlug"))
	if itemSlug == "" {
		writeError(w, r, http.StatusBadRequest, "missing_path_param", "missing item slug")
		return
	}

	from, err := parseTimeQuery(r, "from", true)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid from query value")
		return
	}
	to, err := parseTimeQuery(r, "to", false)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid to query value")
		return
	}

	payload, err := a.store.GetItemChanges(patches.ItemChangesQuery{
		ItemSlug: itemSlug,
		From:     from,
		To:       to,
	})
	if err != nil {
		if errors.Is(err, patches.ErrItemNotFound) {
			writeError(w, r, http.StatusNotFound, "resource_not_found", "item not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load item changes")
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (a *API) listSpells(w http.ResponseWriter, r *http.Request) {
	payload, err := a.store.ListSpells()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list spells")
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *API) getSpellChanges(w http.ResponseWriter, r *http.Request) {
	spellSlug := strings.TrimSpace(chi.URLParam(r, "spellSlug"))
	if spellSlug == "" {
		writeError(w, r, http.StatusBadRequest, "missing_path_param", "missing spell slug")
		return
	}

	from, err := parseTimeQuery(r, "from", true)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid from query value")
		return
	}
	to, err := parseTimeQuery(r, "to", false)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid to query value")
		return
	}

	payload, err := a.store.GetSpellChanges(patches.SpellChangesQuery{
		SpellSlug: spellSlug,
		From:      from,
		To:        to,
	})
	if err != nil {
		if errors.Is(err, patches.ErrSpellNotFound) {
			writeError(w, r, http.StatusNotFound, "resource_not_found", "spell not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load spell changes")
		return
	}

	writeJSON(w, http.StatusOK, payload)
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

func parseTimeQuery(r *http.Request, key string, startOfDay bool) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		utc := parsed.UTC()
		return &utc, nil
	}

	if parsed, err := time.Parse("2006-01-02", raw); err == nil {
		utc := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		if !startOfDay {
			utc = utc.Add(24*time.Hour - time.Nanosecond)
		}
		return &utc, nil
	}

	return nil, errors.New("invalid time query")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type errorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

type errorResponse struct {
	Error errorPayload `json:"error"`
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, errorResponse{
		Error: errorPayload{
			Code:      code,
			Message:   message,
			RequestID: middleware.GetReqID(r.Context()),
		},
	})
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
