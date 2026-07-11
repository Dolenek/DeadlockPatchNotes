package patches

import "testing"

func TestDeduplicateTimelineEvents_PrefersPatchClosestToEvent(t *testing.T) {
	duplicateBlock := PatchTimelineBlock{
		ID:         "shared-event",
		Kind:       "hotfix",
		ReleasedAt: "2026-03-21T12:00:00Z",
		Changes:    []PatchChange{{ID: "change", Text: "Vortex Web cooldown reduced"}},
	}
	details := []PatchDetail{
		{Slug: "03-06-2026-update", PublishedAt: "2026-03-06T12:00:00Z", Timeline: []PatchTimelineBlock{duplicateBlock}},
		{Slug: "03-21-2026-update", PublishedAt: "2026-03-21T19:00:00Z", Timeline: []PatchTimelineBlock{duplicateBlock}},
	}

	deduplicated := deduplicateTimelineEvents(details)
	if len(deduplicated[0].Timeline) != 0 {
		t.Fatalf("expected duplicate event removed from older patch, got %d blocks", len(deduplicated[0].Timeline))
	}
	if len(deduplicated[1].Timeline) != 1 {
		t.Fatalf("expected canonical event retained on closest patch, got %d blocks", len(deduplicated[1].Timeline))
	}
	if len(details[0].Timeline) != 1 {
		t.Fatal("deduplication mutated patch-detail input")
	}
}
