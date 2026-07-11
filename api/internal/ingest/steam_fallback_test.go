package ingest

import (
	"testing"
	"time"
)

func TestAppendSteamMinorUpdatesAddsOnlyNewHotfixes(t *testing.T) {
	patchPublishedAt := time.Date(2026, time.June, 30, 17, 22, 14, 0, time.UTC)
	existing := []timelineCandidate{
		{
			Key:        "initial",
			Kind:       "initial",
			Title:      "Initial Update",
			ReleasedAt: patchPublishedAt,
			BodyText:   "Initial body",
		},
	}
	updates := []SteamMinorUpdate{
		{GID: "initial-copy", PublishedAt: patchPublishedAt, BodyText: "Initial body"},
		{GID: "july-1", SourceURL: "https://example.test/july-1", PublishedAt: time.Date(2026, time.July, 1, 21, 34, 59, 0, time.UTC), BodyText: "Shiv change"},
		{GID: "july-9", SourceURL: "https://example.test/july-9", PublishedAt: time.Date(2026, time.July, 9, 18, 6, 55, 0, time.UTC), BodyText: "Urn change"},
		{GID: "duplicate", PublishedAt: time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC), BodyText: "Urn change"},
	}

	blocks, added, deferred := appendSteamMinorUpdates(existing, updates, patchPublishedAt)
	if deferred != 0 {
		t.Fatalf("expected no deferred updates, got %d", deferred)
	}
	if added != 2 || len(blocks) != 3 {
		t.Fatalf("expected 2 added blocks and 3 total, got added=%d total=%d", added, len(blocks))
	}
	if blocks[0].Kind != "initial" || blocks[1].Kind != "hotfix" || blocks[2].Kind != "hotfix" {
		t.Fatalf("unexpected block kinds: %q, %q, %q", blocks[0].Kind, blocks[1].Kind, blocks[2].Kind)
	}
	if blocks[1].Key != "steam-announcement-july-1" || blocks[2].Key != "steam-announcement-july-9" {
		t.Fatalf("unexpected fallback block keys: %q, %q", blocks[1].Key, blocks[2].Key)
	}
	if blocks[2].Title != "Hotfix 2026-07-09" {
		t.Fatalf("unexpected hotfix title %q", blocks[2].Title)
	}
}

func TestAppendSteamMinorUpdatesDefersDistantNews(t *testing.T) {
	patchPublishedAt := time.Date(2026, time.May, 22, 12, 0, 0, 0, time.UTC)
	existing := []timelineCandidate{{Key: "initial", ReleasedAt: patchPublishedAt, BodyText: "Initial body"}}
	updates := []SteamMinorUpdate{{
		GID:         "new-patch-candidate",
		PublishedAt: patchPublishedAt.Add(20 * 24 * time.Hour),
		BodyText:    "Distant update",
	}}

	blocks, added, deferred := appendSteamMinorUpdates(existing, updates, patchPublishedAt)
	if added != 0 || deferred != 1 || len(blocks) != 1 {
		t.Fatalf("expected distant news to be deferred, got added=%d deferred=%d blocks=%d", added, deferred, len(blocks))
	}
}
