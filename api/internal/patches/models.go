package patches

import (
	"encoding/json"
	"strings"
	"time"
)

// PatchSummary is a compact representation used on the list page.
type PatchSummary struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	PublishedAt   string `json:"publishedAt"`
	Category      string `json:"category"`
	CoverImageURL string `json:"imageUrl"`
	Source        PatchSource `json:"source"`
	Timeline      []PatchTimelineSummary `json:"releaseTimeline,omitempty"`
}

type PatchTimelineSummary struct {
	ID         string `json:"id"`
	Kind       string `json:"releaseType"`
	Title      string `json:"title"`
	ReleasedAt string `json:"releasedAt"`
}

func (s *PatchTimelineSummary) UnmarshalJSON(data []byte) error {
	type alias PatchTimelineSummary
	var payload struct {
		alias
		LegacyKind string `json:"kind"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	*s = PatchTimelineSummary(payload.alias)
	if strings.TrimSpace(s.Kind) == "" {
		s.Kind = strings.TrimSpace(payload.LegacyKind)
	}

	return nil
}

// PatchChange is a single bullet/line under an entry.
type PatchChange struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// PatchEntryGroup is a nested grouping for hero ability/talent blocks.
type PatchEntryGroup struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	IconURL         string      `json:"iconUrl,omitempty"`
	IconFallbackURL string      `json:"iconFallbackUrl,omitempty"`
	Changes         []PatchChange `json:"changes"`
}

// PatchEntry groups related changes.
type PatchEntry struct {
	ID                    string            `json:"id"`
	EntityName            string            `json:"entityName"`
	EntityIconURL         string            `json:"entityIconUrl,omitempty"`
	EntityIconFallbackURL string            `json:"entityIconFallbackUrl,omitempty"`
	Summary               string            `json:"summary,omitempty"`
	Changes               []PatchChange     `json:"changes"`
	Groups                []PatchEntryGroup `json:"groups,omitempty"`
}

// PatchSection is a top-level section in a patch article.
type PatchSection struct {
	ID      string       `json:"id"`
	Title   string       `json:"title"`
	Kind    string       `json:"kind"`
	Entries []PatchEntry `json:"entries"`
}

// PatchSource tracks where the content came from.
type PatchSource struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// PatchTimelineBlock captures initial release and follow-up hotfixes.
type PatchTimelineBlock struct {
	ID         string        `json:"id"`
	Kind       string        `json:"releaseType"`
	Title      string        `json:"title"`
	ReleasedAt string        `json:"releasedAt"`
	Source     PatchSource   `json:"source"`
	Changes    []PatchChange `json:"changes"`
	Sections   []PatchSection `json:"sections,omitempty"`
}

func (b *PatchTimelineBlock) UnmarshalJSON(data []byte) error {
	type alias PatchTimelineBlock
	var payload struct {
		alias
		LegacyKind string `json:"kind"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	*b = PatchTimelineBlock(payload.alias)
	if strings.TrimSpace(b.Kind) == "" {
		b.Kind = strings.TrimSpace(payload.LegacyKind)
	}

	return nil
}

// PatchDetail powers the patch detail page.
type PatchDetail struct {
	ID           string            `json:"id"`
	Slug         string            `json:"slug"`
	Title        string            `json:"title"`
	PublishedAt  string            `json:"publishedAt"`
	Category     string            `json:"category"`
	Source       PatchSource       `json:"source"`
	HeroImageURL string            `json:"imageUrl"`
	Intro        string            `json:"intro"`
	Sections     []PatchSection    `json:"sections"`
	Timeline     []PatchTimelineBlock `json:"releaseTimeline,omitempty"`
}

func (d *PatchDetail) UnmarshalJSON(data []byte) error {
	type alias PatchDetail
	var payload struct {
		alias
		LegacyHeroImageURL string               `json:"heroImageUrl"`
		LegacyTimeline     []PatchTimelineBlock `json:"timeline"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	*d = PatchDetail(payload.alias)
	if strings.TrimSpace(d.HeroImageURL) == "" {
		d.HeroImageURL = strings.TrimSpace(payload.LegacyHeroImageURL)
	}
	if len(d.Timeline) == 0 && len(payload.LegacyTimeline) > 0 {
		d.Timeline = payload.LegacyTimeline
	}

	return nil
}

// listItem stores both summary and detail while preserving a sortable timestamp.
type listItem struct {
	summary   PatchSummary
	detail    PatchDetail
	published time.Time
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// PatchListResponse is the list endpoint payload.
type PatchListResponse struct {
	Patches    []PatchSummary `json:"patches"`
	Pagination Pagination     `json:"pagination"`
}

type HeroSummary struct {
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	IconURL         string `json:"iconUrl,omitempty"`
	IconFallbackURL string `json:"iconFallbackUrl,omitempty"`
	LastChangedAt   string `json:"lastChangedAt"`
}

type HeroListResponse struct {
	Items []HeroSummary `json:"heroes"`
}

type TimelinePatchRef struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type HeroTimelineSkill struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	IconURL         string      `json:"iconUrl,omitempty"`
	IconFallbackURL string      `json:"iconFallbackUrl,omitempty"`
	Changes         []PatchChange `json:"changes"`
}

type HeroTimelineBlock struct {
	ID             string            `json:"id"`
	Kind           string            `json:"releaseType"`
	Label          string            `json:"displayLabel"`
	ReleasedAt     string            `json:"releasedAt"`
	Patch          TimelinePatchRef  `json:"patchRef"`
	Source         PatchSource       `json:"source"`
	GeneralChanges []PatchChange      `json:"generalChanges,omitempty"`
	Skills         []HeroTimelineSkill `json:"skills"`
}

type HeroChangesResponse struct {
	Hero  HeroSummary         `json:"hero"`
	Items []HeroTimelineBlock `json:"timeline"`
}

type HeroChangesQuery struct {
	HeroSlug string
	Skill    string
	From     *time.Time
	To       *time.Time
}

type ItemSummary struct {
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	IconURL         string `json:"iconUrl,omitempty"`
	IconFallbackURL string `json:"iconFallbackUrl,omitempty"`
	LastChangedAt   string `json:"lastChangedAt"`
}

type ItemListResponse struct {
	Items []ItemSummary `json:"items"`
}

type ItemTimelineBlock struct {
	ID         string           `json:"id"`
	Kind       string           `json:"releaseType"`
	Label      string           `json:"displayLabel"`
	ReleasedAt string           `json:"releasedAt"`
	Patch      TimelinePatchRef `json:"patchRef"`
	Source     PatchSource      `json:"source"`
	Changes    []PatchChange    `json:"changes"`
}

type ItemChangesResponse struct {
	Item  ItemSummary        `json:"item"`
	Items []ItemTimelineBlock `json:"timeline"`
}

type ItemChangesQuery struct {
	ItemSlug string
	From     *time.Time
	To       *time.Time
}

type SpellSummary struct {
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	IconURL         string `json:"iconUrl,omitempty"`
	IconFallbackURL string `json:"iconFallbackUrl,omitempty"`
	LastChangedAt   string `json:"lastChangedAt"`
}

type SpellListResponse struct {
	Items []SpellSummary `json:"spells"`
}

type SpellTimelineEntry struct {
	ID              string      `json:"id"`
	HeroSlug        string      `json:"heroSlug,omitempty"`
	HeroName        string      `json:"heroName,omitempty"`
	HeroIconURL     string      `json:"heroIconUrl,omitempty"`
	HeroIconFallbackURL string  `json:"heroIconFallbackUrl,omitempty"`
	Changes         []PatchChange `json:"changes"`
}

type SpellTimelineBlock struct {
	ID         string             `json:"id"`
	Kind       string             `json:"releaseType"`
	Label      string             `json:"displayLabel"`
	ReleasedAt string             `json:"releasedAt"`
	Patch      TimelinePatchRef   `json:"patchRef"`
	Source     PatchSource        `json:"source"`
	Entries    []SpellTimelineEntry `json:"entries"`
}

type SpellChangesResponse struct {
	Spell SpellSummary       `json:"spell"`
	Items []SpellTimelineBlock `json:"timeline"`
}

type SpellChangesQuery struct {
	SpellSlug string
	From      *time.Time
	To        *time.Time
}
