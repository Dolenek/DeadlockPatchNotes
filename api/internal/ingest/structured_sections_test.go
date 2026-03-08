package ingest

import "testing"

func TestBuildStructuredSections_ParsesExplicitSections(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ General ]
- Zipline speed increased
[ Items ]
- Active Reload: Cooldown reduced from 20s to 18s
[ Heroes ]
- Haze: Sleep Dagger Cooldown reduced from 30s to 25s
- Haze: Talents +10% Dagger Radius`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	if sections[0].ID != "general" || sections[1].ID != "items" || sections[2].ID != "heroes" {
		t.Fatalf("unexpected section order: %+v", []string{sections[0].ID, sections[1].ID, sections[2].ID})
	}

	if got := len(sections[0].Entries[0].Changes); got != 1 {
		t.Fatalf("expected 1 general change, got %d", got)
	}
	if got := len(sections[1].Entries); got != 1 {
		t.Fatalf("expected 1 item entry, got %d", got)
	}
	if got := len(sections[2].Entries); got != 1 {
		t.Fatalf("expected 1 hero entry, got %d", got)
	}

	hero := sections[2].Entries[0]
	if hero.EntityName != "Haze" {
		t.Fatalf("expected hero Haze, got %q", hero.EntityName)
	}
	if len(hero.Groups) != 2 {
		t.Fatalf("expected 2 hero groups, got %d", len(hero.Groups))
	}
}

func TestBuildStructuredSections_InfersSectionsWithoutHeaders(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `- Bebop: Stamina increased from 2 to 3
- Active Reload: You can now reload while full
- Base guardian bounty increased`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	if sections[0].ID != "general" || sections[1].ID != "items" || sections[2].ID != "heroes" {
		t.Fatalf("unexpected section order: %+v", []string{sections[0].ID, sections[1].ID, sections[2].ID})
	}

	if sections[2].Entries[0].EntityName != "Bebop" {
		t.Fatalf("expected Bebop hero entry, got %q", sections[2].Entries[0].EntityName)
	}
	if sections[1].Entries[0].EntityName != "Active Reload" {
		t.Fatalf("expected Active Reload item entry, got %q", sections[1].Entries[0].EntityName)
	}
	if len(sections[0].Entries[0].Changes) != 1 {
		t.Fatalf("expected one general change, got %d", len(sections[0].Entries[0].Changes))
	}
}

func testAssetCatalog() *AssetCatalog {
	catalog := &AssetCatalog{
		heroesByNorm: map[string]heroAsset{
			"haze":  {ID: 1, Name: "Haze", Images: heroImages{IconImageSmall: "https://example.test/haze.png"}},
			"bebop": {ID: 2, Name: "Bebop", Images: heroImages{IconImageSmall: "https://example.test/bebop.png"}},
		},
		heroByID: map[int]heroAsset{
			1: {ID: 1, Name: "Haze", Images: heroImages{IconImageSmall: "https://example.test/haze.png"}},
			2: {ID: 2, Name: "Bebop", Images: heroImages{IconImageSmall: "https://example.test/bebop.png"}},
		},
		itemsByNorm: map[string]itemAsset{
			"active reload": {Name: "Active Reload", Type: "item", ShopImage: "https://example.test/active_reload.png"},
		},
		abilitiesByHero: map[string][]abilityRef{
			"haze": {
				{Name: "Sleep Dagger", NormName: "sleep dagger", Image: "https://example.test/sleep_dagger.png"},
			},
		},
	}
	return catalog
}
