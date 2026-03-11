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

func TestHydratePatchDetail_BindsPrefixedAbilityLinesToCurrentHero(t *testing.T) {
	detail := PatchDetail{
		Slug:        "test-update",
		PublishedAt: "2026-03-06T12:00:00Z",
		Sections: []PatchSection{
			{
				ID:    "heroes",
				Title: "Heroes",
				Kind:  "heroes",
				Entries: []PatchEntry{
					{
						ID:         "bebop",
						EntityName: "Bebop",
						Groups: []PatchEntryGroup{
							{ID: "hook", Title: "Hook"},
							{ID: "hyper-beam", Title: "Hyper Beam"},
						},
					},
					{
						ID:         "calico",
						EntityName: "Calico",
						Groups: []PatchEntryGroup{
							{ID: "leaping-slash", Title: "Leaping Slash"},
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
				Changes: []PatchChange{
					{ID: "1", Text: "[Heroes]"},
					{ID: "2", Text: "Bebop"},
					{ID: "3", Text: "Hook: Reworked code to reduce mispredicts."},
					{ID: "4", Text: "Hyper Beam: Effect revisions for projections on vertical surfaces."},
					{ID: "5", Text: "Calico"},
					{ID: "6", Text: "Leaping Slash: Fixed animation getting stuck when stunned during the ability cast."},
				},
			},
		},
	}

	hydrated := hydratePatchDetail(detail)
	heroes := timelineSectionByKind(hydrated.Timeline[0].Sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	if len(heroes.Entries) != 2 {
		t.Fatalf("expected 2 hero entries, got %d", len(heroes.Entries))
	}

	bebop := timelineEntryByName(heroes.Entries, "Bebop")
	if bebop == nil {
		t.Fatal("expected Bebop entry")
	}
	if timelineGroupByTitle(*bebop, "Hook") == nil {
		t.Fatal("expected Hook group under Bebop")
	}
	if timelineGroupByTitle(*bebop, "Hyper Beam") == nil {
		t.Fatal("expected Hyper Beam group under Bebop")
	}

	if timelineEntryByName(heroes.Entries, "Hook") != nil || timelineEntryByName(heroes.Entries, "Hyper Beam") != nil {
		t.Fatal("ability prefixes should not become standalone hero entries")
	}
}

func TestHydratePatchDetail_CardTypesRemainUnderHeroGroup(t *testing.T) {
	detail := PatchDetail{
		Slug:        "test-update",
		PublishedAt: "2026-03-06T12:00:00Z",
		Sections: []PatchSection{
			{
				ID:    "heroes",
				Title: "Heroes",
				Kind:  "heroes",
				Entries: []PatchEntry{
					{
						ID:         "wraith",
						EntityName: "Wraith",
						Groups: []PatchEntryGroup{
							{ID: "card-trick", Title: "Card Trick"},
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
				Changes: []PatchChange{
					{ID: "1", Text: "[Heroes]"},
					{ID: "2", Text: "Wraith"},
					{ID: "3", Text: "Card Trick cards now have specific suites with special bonuses."},
					{ID: "4", Text: "Card Types:"},
					{ID: "5", Text: "Spades: +70% Damage"},
					{ID: "6", Text: "Diamond: Cuts enemy resistances by -8% for 5s."},
				},
			},
		},
	}

	hydrated := hydratePatchDetail(detail)
	heroes := timelineSectionByKind(hydrated.Timeline[0].Sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	if len(heroes.Entries) != 1 {
		t.Fatalf("expected 1 hero entry, got %d", len(heroes.Entries))
	}

	cardTypes := timelineGroupByTitle(heroes.Entries[0], "Card Types")
	if cardTypes == nil {
		t.Fatal("expected Card Types group")
	}
	if len(cardTypes.Changes) != 2 {
		t.Fatalf("expected 2 card type changes, got %d", len(cardTypes.Changes))
	}

	if timelineEntryByName(heroes.Entries, "Spades") != nil || timelineEntryByName(heroes.Entries, "Diamond") != nil {
		t.Fatal("card type labels should not become standalone hero entries")
	}
}

func timelineSectionByKind(sections []PatchSection, kind string) *PatchSection {
	for i := range sections {
		if sections[i].Kind == kind {
			return &sections[i]
		}
	}
	return nil
}

func timelineEntryByName(entries []PatchEntry, name string) *PatchEntry {
	for i := range entries {
		if entries[i].EntityName == name {
			return &entries[i]
		}
	}
	return nil
}

func timelineGroupByTitle(entry PatchEntry, title string) *PatchEntryGroup {
	for i := range entry.Groups {
		if entry.Groups[i].Title == title {
			return &entry.Groups[i]
		}
	}
	return nil
}
