package ingest

import (
	"testing"
	"time"
)

func TestStoredPatchStateMatchesSemanticJSON(t *testing.T) {
	timestamp := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	left := storedPatchState{
		Slug:          "update",
		Title:         "Update",
		PublishedAt:   timestamp,
		UpdatedAt:     timestamp,
		DetailPayload: []byte(`{"title":"Update","sections":[]}`),
	}
	right := left
	right.DetailPayload = []byte(`{"sections": [], "title": "Update"}`)

	if !left.matches(right) {
		t.Fatal("semantically identical JSON payloads should match")
	}
	right.Title = "Changed"
	if left.matches(right) {
		t.Fatal("changed patch fields should not match")
	}
}

func TestTimelineCandidatesEqualChecksStoredSourceContent(t *testing.T) {
	timestamp := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	left := timelineCandidate{Key: "initial", Kind: "initial", BodyText: "change", ReleasedAt: timestamp}
	right := left
	if !timelineCandidatesEqual(left, right) {
		t.Fatal("identical timeline candidates should match")
	}
	right.BodyText = "different"
	if timelineCandidatesEqual(left, right) {
		t.Fatal("changed source body should not match")
	}
}
