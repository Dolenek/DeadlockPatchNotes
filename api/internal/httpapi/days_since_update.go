package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var errLastUpdateHeroNotFound = errors.New("hero not found")
var errLastUpdatePatchNotFound = errors.New("patch not found")

func (a *API) daysSinceLastUpdate(w http.ResponseWriter, r *http.Request) {
	heroSlug := strings.TrimSpace(r.URL.Query().Get("hero"))
	onlyUpdate, err := parseOnlyUpdateQuery(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query_param", "invalid onlyUpdate query value")
		return
	}

	baseline, err := a.resolveDaysSinceBaseline(heroSlug, onlyUpdate)
	if err != nil {
		writeDaysSinceBaselineError(w, r, err)
		return
	}

	location, err := resolveBerlinLocation()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve feed timezone")
		return
	}

	days := daysSinceLastUpdate(baseline, rssNow().UTC(), location)
	writePlainText(w, http.StatusOK, fmt.Sprintf("Days since last update: %d", days))
}

func (a *API) resolveDaysSinceBaseline(heroSlug string, onlyUpdate bool) (time.Time, error) {
	lastPatchAt, err := a.resolveLatestPatchTimestamp()
	if err != nil {
		return time.Time{}, err
	}
	if heroSlug == "" {
		return lastPatchAt, nil
	}

	heroLastUpdateAt, err := a.resolveHeroLastChangedTimestamp(heroSlug)
	if err != nil {
		return time.Time{}, err
	}
	return resolveDaysBaseline(heroLastUpdateAt, lastPatchAt, onlyUpdate), nil
}

func writeDaysSinceBaselineError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errLastUpdateHeroNotFound):
		writeError(w, r, http.StatusNotFound, "resource_not_found", "hero not found")
	case errors.Is(err, errLastUpdatePatchNotFound):
		writeError(w, r, http.StatusNotFound, "resource_not_found", "no patches found")
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve days-since baseline")
	}
}

func resolveDaysBaseline(lastHeroUpdateAt, lastPatchAt time.Time, onlyUpdate bool) time.Time {
	if onlyUpdate {
		return lastHeroUpdateAt
	}
	if lastPatchAt.After(lastHeroUpdateAt) {
		return lastPatchAt
	}
	return lastHeroUpdateAt
}

func (a *API) resolveLatestPatchTimestamp() (time.Time, error) {
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

func (a *API) resolveHeroLastChangedTimestamp(heroSlug string) (time.Time, error) {
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

func parseOnlyUpdateQuery(r *http.Request) (bool, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("onlyUpdate"))
	if raw == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, err
	}
	return parsed, nil
}

func writePlainText(w http.ResponseWriter, status int, value string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(value))
}
