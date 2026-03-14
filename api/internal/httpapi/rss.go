package httpapi

import (
	"encoding/xml"
	"errors"
	"net/http"
	"time"

	"deadlockpatchnotes/api/internal/patches"
	"github.com/go-chi/chi/v5"
)

const rssMaxItems = 50

var rssNow = time.Now

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	AtomNS  string     `xml:"xmlns:atom,attr,omitempty"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string      `xml:"title"`
	Link          string      `xml:"link"`
	Description   string      `xml:"description"`
	Language      string      `xml:"language,omitempty"`
	LastBuildDate string      `xml:"lastBuildDate,omitempty"`
	AtomLink      rssAtomLink `xml:"atom:link"`
	Items         []rssItem   `xml:"item"`
}

type rssAtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	GUID        rssGUID `xml:"guid"`
	Description string  `xml:"description"`
	PubDate     string  `xml:"pubDate"`
}

type rssGUID struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

func (a *API) patchFeedRSS(w http.ResponseWriter, r *http.Request) {
	patchesPayload, err := listAllPatchSummaries(a.store)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list patches")
		return
	}
	patchesPayload = preparePatchSummariesForFeed(patchesPayload, rssMaxItems)

	siteBaseURL := resolveFeedSiteBaseURL(r)
	now := rssNow().UTC()
	items, err := buildPatchRSSItems(a.store, patchesPayload, siteBaseURL, now)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load patch feed item")
		return
	}

	doc := rssDocument{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         "Deadlock Patch Notes - Patch Updates",
			Link:          buildAbsoluteURL(siteBaseURL, "/patches"),
			Description:   "Latest Deadlock patch updates as they land in the archive.",
			Language:      "en-us",
			LastBuildDate: now.Format(time.RFC1123Z),
			AtomLink: rssAtomLink{
				Href: resolveRequestURL(r),
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeRSS(w, http.StatusOK, doc)
}

func (a *API) heroFeedRSS(w http.ResponseWriter, r *http.Request) {
	heroSlug := strings.TrimSpace(chi.URLParam(r, "heroSlug"))
	if heroSlug == "" {
		writeError(w, r, http.StatusBadRequest, "missing_path_param", "missing hero slug")
		return
	}

	payload, err := a.store.GetHeroChanges(patches.HeroChangesQuery{HeroSlug: heroSlug})
	if err != nil {
		if errors.Is(err, patches.ErrHeroNotFound) {
			writeError(w, r, http.StatusNotFound, "resource_not_found", "hero not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load hero feed")
		return
	}

	grouped := groupHeroTimelineByPatch(payload.Items)
	if len(grouped) > rssMaxItems {
		grouped = grouped[:rssMaxItems]
	}

	siteBaseURL := resolveFeedSiteBaseURL(r)
	now := rssNow().UTC()
	items, heroLink, heroName := buildHeroRSSItems(payload.Hero, grouped, siteBaseURL)

	doc := newHeroRSSDocument(heroName, heroLink, resolveRequestURL(r), items, now)

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeRSS(w, http.StatusOK, doc)
}

func (a *API) heroDaysWithoutUpdateRSS(w http.ResponseWriter, r *http.Request) {
	heroSlug := strings.TrimSpace(chi.URLParam(r, "heroSlug"))
	if heroSlug == "" {
		writeError(w, r, http.StatusBadRequest, "missing_path_param", "missing hero slug")
		return
	}

	heroesPayload, err := a.store.ListHeroes()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load heroes")
		return
	}

	hero, found := findHeroBySlug(heroesPayload.Items, heroSlug)
	if !found {
		writeError(w, r, http.StatusNotFound, "resource_not_found", "hero not found")
		return
	}

	location, err := resolveBerlinLocation()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve feed timezone")
		return
	}

	now := rssNow().UTC()
	lastChanged := parseTimeRFC3339(hero.LastChangedAt)
	daysWithoutUpdate := daysSinceLastUpdate(lastChanged, now, location)
	item, heroLink := buildHeroDaysWithoutUpdateRSSItem(hero, daysWithoutUpdate, lastChanged, location, resolveFeedSiteBaseURL(r), now)

	doc := newHeroDaysWithoutUpdateRSSDocument(hero.Name, heroLink, resolveRequestURL(r), item, now)

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeRSS(w, http.StatusOK, doc)
}

func writeRSS(w http.ResponseWriter, status int, document rssDocument) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	_ = encoder.Encode(document)
}
