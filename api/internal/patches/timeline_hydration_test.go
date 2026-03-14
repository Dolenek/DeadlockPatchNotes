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
							{ID: "grapple-arm", Title: "Grapple Arm"},
							{ID: "hyper-beam", Title: "Hyper Beam"},
							{ID: "exploding-uppercut", Title: "Exploding Uppercut"},
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
					{ID: "4", Text: "Hyperbeam: Effect revisions for projections on vertical surfaces."},
					{ID: "5", Text: "Uppercut: T3 no longer grants +100% Ammo"},
					{ID: "6", Text: "Calico"},
					{ID: "7", Text: "Leaping Slash: Fixed animation getting stuck when stunned during the ability cast."},
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
	if timelineGroupByTitle(*bebop, "Grapple Arm") == nil {
		t.Fatal("expected Grapple Arm group under Bebop")
	}
	if timelineGroupByTitle(*bebop, "Hyper Beam") == nil {
		t.Fatal("expected Hyper Beam group under Bebop")
	}
	if timelineGroupByTitle(*bebop, "Exploding Uppercut") == nil {
		t.Fatal("expected Exploding Uppercut group under Bebop")
	}

	if timelineEntryByName(heroes.Entries, "Hook") != nil || timelineEntryByName(heroes.Entries, "Hyperbeam") != nil || timelineEntryByName(heroes.Entries, "Uppercut") != nil {
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

func TestHydratePatchDetail_CanonicalizesDoormanNameAcrossArticleVariants(t *testing.T) {
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
						ID:         "the-doorman",
						EntityName: "The Doorman",
						Groups: []PatchEntryGroup{
							{ID: "call-bell", Title: "Call Bell"},
							{ID: "hotel-guest", Title: "Hotel Guest"},
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
					{ID: "2", Text: "The Doorman: Call Bell cooldown increased from 16s to 18s."},
					{ID: "3", Text: "Doorman: Hotel Guest cast range increased from 6m to 7m."},
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
	if heroes.Entries[0].EntityName != "Doorman" {
		t.Fatalf("expected canonical Doorman name, got %q", heroes.Entries[0].EntityName)
	}
	if timelineGroupByTitle(heroes.Entries[0], "Call Bell") == nil {
		t.Fatal("expected Call Bell group under Doorman")
	}
	if timelineGroupByTitle(heroes.Entries[0], "Hotel Guest") == nil {
		t.Fatal("expected Hotel Guest group under Doorman")
	}
}

func TestHydratePatchDetail_KeepsDoormanFollowupAbilityLinesOutOfGeneral(t *testing.T) {
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
						ID:         "the-doorman",
						EntityName: "The Doorman",
						Groups: []PatchEntryGroup{
							{ID: "call-bell", Title: "Call Bell"},
							{ID: "doorway", Title: "Doorway"},
							{ID: "luggage-cart", Title: "Luggage Cart"},
							{ID: "hotel-guest", Title: "Hotel Guest"},
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
					{ID: "2", Text: "Doorman"},
					{ID: "3", Text: "Gun now pierces through targets at 50% reduced damage"},
					{ID: "4", Text: "Call Bell time between charges increased from 4s to 6s"},
					{ID: "5", Text: "Doorway now has a timer icon above the ability"},
					{ID: "6", Text: "Luggage Cart is now 20% larger (20% wider hitbox as well)"},
					{ID: "7", Text: "Hotel Guest cast range increased from 6m to 7m"},
				},
			},
		},
	}

	hydrated := hydratePatchDetail(detail)
	heroes := timelineSectionByKind(hydrated.Timeline[0].Sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	doorman := timelineEntryByName(heroes.Entries, "Doorman")
	if doorman == nil {
		t.Fatal("expected Doorman entry")
	}
	if len(doorman.Changes) != 1 {
		t.Fatalf("expected 1 Doorman general change, got %d", len(doorman.Changes))
	}
	if doorman.Changes[0].Text != "Gun now pierces through targets at 50% reduced damage" {
		t.Fatalf("unexpected Doorman general change: %+v", doorman.Changes)
	}

	for _, title := range []string{"Call Bell", "Doorway", "Luggage Cart", "Hotel Guest"} {
		if timelineGroupByTitle(*doorman, title) == nil {
			t.Fatalf("expected %s group under Doorman", title)
		}
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
