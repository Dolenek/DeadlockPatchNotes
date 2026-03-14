package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var errLastUpdateHeroNotFound = errors.New("hero not found")
var errLastUpdatePatchNotFound = errors.New("patch not found")

func (a *API) daysSinceLastUpdate(w http.ResponseWriter, r *http.Request) {
	lastUpdateAt, err := a.resolveLastUpdateTimestamp(strings.TrimSpace(r.URL.Query().Get("hero")))
	if err != nil {
		switch {
		case errors.Is(err, errLastUpdateHeroNotFound):
			writeError(w, r, http.StatusNotFound, "resource_not_found", "hero not found")
		case errors.Is(err, errLastUpdatePatchNotFound):
			writeError(w, r, http.StatusNotFound, "resource_not_found", "no patches found")
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve last update timestamp")
		}
		return
	}

	location, err := resolveBerlinLocation()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve feed timezone")
		return
	}

	days := daysSinceLastUpdate(lastUpdateAt, rssNow().UTC(), location)
	writePlainText(w, http.StatusOK, fmt.Sprintf("Days since last update: %d", days))
}

func (a *API) resolveLastUpdateTimestamp(heroSlug string) (time.Time, error) {
	if heroSlug != "" {
		heroesPayload, err := a.store.ListHeroes()
		if err != nil {
			return time.Time{}, err
		}
		hero, found := findHeroBySlug(heroesPayload.Items, heroSlug)
		if !found {
			return time.Time{}, errLastUpdateHeroNotFound
		}
		return parseTimeRFC3339(hero.LastChangedAt), nil
	}

	patchSummaries, err := listAllPatchSummaries(a.store)
	if err != nil {
		return time.Time{}, err
	}
	latestPatch := preparePatchSummariesForFeed(patchSummaries, 1)
	if len(latestPatch) == 0 {
		return time.Time{}, errLastUpdatePatchNotFound
	}
	return parseTimeRFC3339(latestPatch[0].PublishedAt), nil
}

func writePlainText(w http.ResponseWriter, status int, value string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(value))
}
