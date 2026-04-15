package patches

import (
	"testing"
	"time"
)

func TestBuildItemList_ReturnsItemsWithChanges(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Timeline: []PatchTimelineBlock{
				{
					ID:         "b1",
					Kind:       "initial",
					ReleasedAt: "2026-03-06T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "items",
							Title: "Items",
							Kind:  "items",
							Entries: []PatchEntry{
								{
									ID:         "active-reload",
									EntityName: "Active Reload",
									Changes:    []PatchChange{{ID: "c1", Text: "Cooldown reduced"}},
								},
							},
						},
					},
				},
			},
		},
	}

	payload := buildItemList(details)
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}
	if payload.Items[0].Slug != "active-reload" {
		t.Fatalf("expected active-reload slug, got %q", payload.Items[0].Slug)
	}
}

func TestBuildItemChanges_AppliesDateFilters(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Timeline: []PatchTimelineBlock{
				{
					ID:         "b1",
					Kind:       "initial",
					ReleasedAt: "2026-03-06T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "items",
							Title: "Items",
							Kind:  "items",
							Entries: []PatchEntry{
								{
									ID:         "active-reload",
									EntityName: "Active Reload",
									Changes:    []PatchChange{{ID: "c1", Text: "Cooldown reduced"}},
								},
							},
						},
					},
				},
				{
					ID:         "b2",
					Kind:       "hotfix",
					ReleasedAt: "2026-03-10T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "items",
							Title: "Items",
							Kind:  "items",
							Entries: []PatchEntry{
								{
									ID:         "active-reload",
									EntityName: "Active Reload",
									Changes:    []PatchChange{{ID: "c2", Text: "Damage increased"}},
								},
							},
						},
					},
				},
			},
		},
	}

	from := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.March, 8, 23, 59, 59, 0, time.UTC)
	payload, err := buildItemChanges(details, ItemChangesQuery{
		ItemSlug: "active-reload",
		From:     &from,
		To:       &to,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 filtered item block, got %d", len(payload.Items))
	}
}

func TestBuildSpellList_SkipsTalentsAndKnownItems(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Timeline: []PatchTimelineBlock{
				{
					ID:         "b1",
					Kind:       "initial",
					ReleasedAt: "2026-03-06T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "items",
							Title: "Items",
							Kind:  "items",
							Entries: []PatchEntry{
								{
									ID:         "active-reload",
									EntityName: "Active Reload",
									Changes:    []PatchChange{{ID: "i1", Text: "Reload speed increased"}},
								},
							},
						},
						{
							ID:    "heroes",
							Title: "Heroes",
							Kind:  "heroes",
							Entries: []PatchEntry{
								{
									ID:                    "abrams",
									EntityName:            "Abrams",
									EntityIconFallbackURL: "https://example.test/abrams.png",
									Groups: []PatchEntryGroup{
										{
											ID:      "shoulder-charge",
											Title:   "Shoulder Charge",
											Changes: []PatchChange{{ID: "s1", Text: "Cooldown reduced"}},
										},
										{
											ID:      "talents",
											Title:   "Talents",
											Changes: []PatchChange{{ID: "t1", Text: "+5% damage"}},
										},
									},
								},
								{
									ID:         "active-reload-polluted",
									EntityName: "Active Reload",
									Changes:    []PatchChange{{ID: "p1", Text: "Should not be spell"}},
								},
								{
									ID:         "base-changes-polluted",
									EntityName: "Base Changes",
									Changes:    []PatchChange{{ID: "p2", Text: "Should not be spell"}},
								},
							},
						},
					},
				},
			},
		},
	}

	payload := buildSpellList(details)
	if len(payload.Items) != 1 {
		t.Fatalf("expected only shoulder-charge spell, got %d", len(payload.Items))
	}
	if payload.Items[0].Slug != "shoulder-charge" {
		t.Fatalf("expected shoulder-charge slug, got %q", payload.Items[0].Slug)
	}
}

func TestBuildSpellChanges_RejectsIconlessStandaloneEntries(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Timeline: []PatchTimelineBlock{
				{
					ID:         "b1",
					Kind:       "initial",
					ReleasedAt: "2026-03-06T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "heroes",
							Title: "Heroes",
							Kind:  "heroes",
							Entries: []PatchEntry{
								{
									ID:         "base-changes-polluted",
									EntityName: "Base Changes",
									Changes:    []PatchChange{{ID: "p1", Text: "Should not be spell"}},
								},
							},
						},
					},
				},
			},
		},
	}

	if _, err := buildSpellChanges(details, SpellChangesQuery{SpellSlug: "base-changes"}); err != ErrSpellNotFound {
		t.Fatalf("expected ErrSpellNotFound, got %v", err)
	}
}

func TestBuildSpellChanges_MergesNameCollisionsAcrossHeroes(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Timeline: []PatchTimelineBlock{
				{
					ID:         "b1",
					Kind:       "initial",
					ReleasedAt: "2026-03-06T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "heroes",
							Title: "Heroes",
							Kind:  "heroes",
							Entries: []PatchEntry{
								{
									ID:                    "abrams",
									EntityName:            "Abrams",
									EntityIconFallbackURL: "https://example.test/abrams.png",
									Groups: []PatchEntryGroup{
										{
											ID:      "shoulder-charge",
											Title:   "Shoulder Charge",
											Changes: []PatchChange{{ID: "a1", Text: "Cooldown reduced"}},
										},
									},
								},
								{
									ID:                    "bebop",
									EntityName:            "Bebop",
									EntityIconFallbackURL: "https://example.test/bebop.png",
									Groups: []PatchEntryGroup{
										{
											ID:      "shoulder-charge",
											Title:   "Shoulder Charge",
											Changes: []PatchChange{{ID: "b1", Text: "Damage increased"}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	payload, err := buildSpellChanges(details, SpellChangesQuery{SpellSlug: "shoulder-charge"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 timeline block, got %d", len(payload.Items))
	}
	if len(payload.Items[0].Entries) != 2 {
		t.Fatalf("expected merged entries from 2 heroes, got %d", len(payload.Items[0].Entries))
	}
}

func TestBuildSpellChanges_IndexesDoormanFollowupAbilityLines(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Sections: []PatchSection{
				{
					ID:    "heroes",
					Title: "Heroes",
					Kind:  "heroes",
					Entries: []PatchEntry{
						{
							ID:                    "the-doorman",
							EntityName:            "The Doorman",
							EntityIconFallbackURL: "https://example.test/doorman.png",
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
					ID:         "b1",
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
		},
	}

	for _, spellSlug := range []string{"call-bell", "doorway", "luggage-cart", "hotel-guest"} {
		payload, err := buildSpellChanges(details, SpellChangesQuery{SpellSlug: spellSlug})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", spellSlug, err)
		}
		if len(payload.Items) != 1 {
			t.Fatalf("expected 1 timeline block for %s, got %d", spellSlug, len(payload.Items))
		}
		if len(payload.Items[0].Entries) != 1 {
			t.Fatalf("expected 1 Doorman entry for %s, got %d", spellSlug, len(payload.Items[0].Entries))
		}
		if payload.Items[0].Entries[0].HeroName != "Doorman" {
			t.Fatalf("expected Doorman hero name for %s, got %q", spellSlug, payload.Items[0].Entries[0].HeroName)
		}
	}
}

func TestBuildEntityQueries_TreatPromotedAbilitiesAsSpellsNotItems(t *testing.T) {
	details := []PatchDetail{
		{
			Slug:  "u1",
			Title: "U1",
			Timeline: []PatchTimelineBlock{
				{
					ID:         "b1",
					Kind:       "initial",
					ReleasedAt: "2026-03-06T12:00:00Z",
					Sections: []PatchSection{
						{
							ID:    "items",
							Title: "Items",
							Kind:  "items",
							Entries: []PatchEntry{
								{
									ID:         "active-reload",
									EntityName: "Active Reload",
									Changes:    []PatchChange{{ID: "i1", Text: "Reload speed increased"}},
								},
							},
						},
						{
							ID:    "heroes",
							Title: "Heroes",
							Kind:  "heroes",
							Entries: []PatchEntry{
								{
									ID:                    "pocket",
									EntityName:            "Pocket",
									EntityIconFallbackURL: "https://example.test/pocket.png",
									Groups: []PatchEntryGroup{
										{
											ID:              "affliction",
											Title:           "Affliction",
											IconFallbackURL: "https://example.test/affliction.png",
											Changes:         []PatchChange{{ID: "s1", Text: "Duration reduced from 18s to 14s"}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	itemsPayload := buildItemList(details)
	if len(itemsPayload.Items) != 1 || itemsPayload.Items[0].Slug != "active-reload" {
		t.Fatalf("expected only Active Reload in items list, got %+v", itemsPayload.Items)
	}

	if _, err := buildItemChanges(details, ItemChangesQuery{ItemSlug: "affliction"}); err != ErrItemNotFound {
		t.Fatalf("expected ErrItemNotFound for affliction item lookup, got %v", err)
	}

	spellsPayload := buildSpellList(details)
	if len(spellsPayload.Items) != 1 || spellsPayload.Items[0].Slug != "affliction" {
		t.Fatalf("expected Affliction spell in spell list, got %+v", spellsPayload.Items)
	}

	spellTimeline, err := buildSpellChanges(details, SpellChangesQuery{SpellSlug: "affliction"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spellTimeline.Items) != 1 || len(spellTimeline.Items[0].Entries) != 1 {
		t.Fatalf("expected one Affliction timeline entry, got %+v", spellTimeline.Items)
	}
	if spellTimeline.Items[0].Entries[0].HeroName != "Pocket" {
		t.Fatalf("expected Pocket Affliction timeline entry, got %+v", spellTimeline.Items[0].Entries[0])
	}
}
