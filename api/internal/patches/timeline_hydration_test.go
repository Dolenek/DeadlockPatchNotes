package patches

import "testing"

func TestHydratePatchDetail_DedupesWhitespaceVariantTimelineBlocks(t *testing.T) {
	detail := PatchDetail{
		Slug:        "test-update",
		PublishedAt: "2026-03-06T12:00:00Z",
		Source: PatchSource{
			Type: "forum-post",
			URL:  "https://example.test/post",
		},
		Sections: []PatchSection{
			{
				ID:    "heroes",
				Title: "Heroes",
				Kind:  "heroes",
				Entries: []PatchEntry{
					{
						ID:         "abrams",
						EntityName: "Abrams",
						Groups: []PatchEntryGroup{
							{ID: "abrams-shoulder-charge", Title: "Shoulder Charge"},
						},
					},
				},
			},
		},
		Timeline: []PatchTimelineBlock{
			{
				ID:         "block-1",
				Kind:       "initial",
				ReleasedAt: "2026-03-06T12:00:00Z",
				Source:     PatchSource{Type: "forum-post", URL: "https://example.test/post-1"},
				Changes: []PatchChange{
					{ID: "1", Text: "Abrams: Shoulder Charge cooldown reduced from 37s to 33s"},
				},
			},
			{
				ID:         "block-2",
				Kind:       "hotfix",
				ReleasedAt: "2026-03-07T12:00:00Z",
				Source:     PatchSource{Type: "forum-post", URL: "https://example.test/post-2"},
				Changes: []PatchChange{
					{ID: "2", Text: "  Abrams:   Shoulder Charge cooldown reduced from 37s to 33s  "},
				},
			},
		},
	}

	hydrated := hydratePatchDetail(detail)
	if len(hydrated.Timeline) != 1 {
		t.Fatalf("expected 1 canonical timeline block, got %d", len(hydrated.Timeline))
	}
	if len(hydrated.Timeline[0].Sections) == 0 {
		t.Fatal("expected hydrated block sections")
	}
}

func TestHydratePatchDetail_SynthesizesTimelineWhenMissing(t *testing.T) {
	detail := PatchDetail{
		Slug:        "test-update",
		PublishedAt: "2026-03-06T12:00:00Z",
		Source: PatchSource{
			Type: "forum-post",
			URL:  "https://example.test/post",
		},
		Sections: []PatchSection{
			{
				ID:    "general",
				Title: "General",
				Kind:  "general",
				Entries: []PatchEntry{
					{
						ID:         "general-gameplay",
						EntityName: "Core Gameplay",
						Changes:    []PatchChange{{ID: "c1", Text: "Zipline speed increased"}},
					},
				},
			},
		},
	}

	hydrated := hydratePatchDetail(detail)
	if len(hydrated.Timeline) != 1 {
		t.Fatalf("expected synthesized timeline with 1 block, got %d", len(hydrated.Timeline))
	}
	if hydrated.Timeline[0].Kind != "initial" {
		t.Fatalf("expected initial synthesized block, got %q", hydrated.Timeline[0].Kind)
	}
}
