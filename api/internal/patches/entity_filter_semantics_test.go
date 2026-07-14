package patches

import (
	"testing"
	"time"
)

func TestEntityChangesKeepMetadataWhenDateRangeHasNoTimelineEntries(t *testing.T) {
	const releasedAt = "2026-03-06T12:00:00Z"
	details := entityFilterFixture(releasedAt)
	future := time.Date(2099, time.January, 1, 0, 0, 0, 0, time.UTC)

	hero, err := buildHeroChanges(details, HeroChangesQuery{HeroSlug: "abrams", From: &future})
	if err != nil {
		t.Fatalf("known hero with empty range returned error: %v", err)
	}
	assertEmptyEntityTimeline(t, "hero", hero.Hero.Name, hero.Hero.LastChangedAt, releasedAt, len(hero.Items))

	item, err := buildItemChanges(details, ItemChangesQuery{ItemSlug: "active-reload", From: &future})
	if err != nil {
		t.Fatalf("known item with empty range returned error: %v", err)
	}
	assertEmptyEntityTimeline(t, "item", item.Item.Name, item.Item.LastChangedAt, releasedAt, len(item.Items))

	spell, err := buildSpellChanges(details, SpellChangesQuery{SpellSlug: "shoulder-charge", From: &future})
	if err != nil {
		t.Fatalf("known spell with empty range returned error: %v", err)
	}
	assertEmptyEntityTimeline(t, "spell", spell.Spell.Name, spell.Spell.LastChangedAt, releasedAt, len(spell.Items))
}

func entityFilterFixture(releasedAt string) []PatchDetail {
	return []PatchDetail{{
		Slug:  "u1",
		Title: "U1",
		Timeline: []PatchTimelineBlock{{
			ID:         "b1",
			Kind:       "initial",
			ReleasedAt: releasedAt,
			Sections: []PatchSection{
				{
					Kind: "heroes",
					Entries: []PatchEntry{{
						EntityName:            "Abrams",
						EntityIconFallbackURL: "https://example.test/abrams.png",
						Changes:               []PatchChange{{ID: "hero-change", Text: "Health increased"}},
						Groups: []PatchEntryGroup{{
							ID:      "shoulder-charge",
							Title:   "Shoulder Charge",
							Changes: []PatchChange{{ID: "spell-change", Text: "Cooldown reduced"}},
						}},
					}},
				},
				{
					Kind: "items",
					Entries: []PatchEntry{{
						EntityName:            "Active Reload",
						EntityIconFallbackURL: "https://example.test/active-reload.png",
						Changes:               []PatchChange{{ID: "item-change", Text: "Cooldown reduced"}},
					}},
				},
			},
		}},
	}}
}

func assertEmptyEntityTimeline(t *testing.T, entityType, name, changedAt, expectedChangedAt string, timelineLength int) {
	t.Helper()
	if name == "" {
		t.Fatalf("%s metadata is empty", entityType)
	}
	if changedAt != expectedChangedAt {
		t.Fatalf("%s lastChangedAt = %q, want %q", entityType, changedAt, expectedChangedAt)
	}
	if timelineLength != 0 {
		t.Fatalf("%s timeline has %d entries, want 0", entityType, timelineLength)
	}
}
