package patches

import "time"

// PatchSummary is a compact representation used on the list page.
type PatchSummary struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	PublishedAt   string `json:"publishedAt"`
	Category      string `json:"category"`
	Excerpt       string `json:"excerpt"`
	CoverImageURL string `json:"coverImageUrl"`
	SourceURL     string `json:"sourceUrl"`
}

// PatchChange is a single bullet/line under an entry.
type PatchChange struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// PatchEntryGroup is a nested grouping for hero ability/talent blocks.
type PatchEntryGroup struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	IconURL         string       `json:"iconUrl,omitempty"`
	IconFallbackURL string       `json:"iconFallbackUrl,omitempty"`
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
	Kind       string        `json:"kind"`
	Title      string        `json:"title"`
	ReleasedAt string        `json:"releasedAt"`
	Source     PatchSource   `json:"source"`
	Changes    []PatchChange `json:"changes"`
}

// PatchDetail powers the patch detail page.
type PatchDetail struct {
	ID           string              `json:"id"`
	Slug         string              `json:"slug"`
	Title        string              `json:"title"`
	PublishedAt  string              `json:"publishedAt"`
	Category     string              `json:"category"`
	Source       PatchSource         `json:"source"`
	HeroImageURL string              `json:"heroImageUrl"`
	Intro        string              `json:"intro"`
	Sections     []PatchSection      `json:"sections"`
	Timeline     []PatchTimelineBlock `json:"timeline,omitempty"`
}

// listItem stores both summary and detail while preserving a sortable timestamp.
type listItem struct {
	summary   PatchSummary
	detail    PatchDetail
	published time.Time
}

// ListResponse is the list endpoint payload.
type ListResponse struct {
	Items      []PatchSummary `json:"items"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int            `json:"total"`
	TotalPages int            `json:"totalPages"`
}
