package ingest

import (
	"testing"

	"deadlockpatchnotes/api/internal/patches"
)

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

func TestBuildStructuredSections_AbilityPrefixesStayOnCurrentHero(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Heroes ]
- Bebop
- Hook: Reworked code to reduce mispredicts.
- Hyper Beam: Effect revisions for projections on vertical surfaces.
- Uppercut: T3 no longer grants +100% Ammo
- Calico
- Leaping Slash: Fixed animation getting stuck when stunned during the ability cast.`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	if len(heroes.Entries) != 2 {
		t.Fatalf("expected 2 hero entries, got %d", len(heroes.Entries))
	}

	bebop := heroByName(heroes.Entries, "Bebop")
	if bebop == nil {
		t.Fatal("expected Bebop entry")
	}
	if groupByTitle(*bebop, "Grapple Arm") == nil {
		t.Fatal("expected Grapple Arm group under Bebop")
	}
	if groupByTitle(*bebop, "Hyper Beam") == nil {
		t.Fatal("expected Hyper Beam group under Bebop")
	}
	if groupByTitle(*bebop, "Exploding Uppercut") == nil {
		t.Fatal("expected Exploding Uppercut group under Bebop")
	}

	calico := heroByName(heroes.Entries, "Calico")
	if calico == nil {
		t.Fatal("expected Calico entry")
	}
	if groupByTitle(*calico, "Leaping Slash") == nil {
		t.Fatal("expected Leaping Slash group under Calico")
	}

	if heroByName(heroes.Entries, "Hook") != nil || heroByName(heroes.Entries, "Hyper Beam") != nil || heroByName(heroes.Entries, "Uppercut") != nil {
		t.Fatal("ability prefixes should not become standalone hero entries")
	}
}

func TestBuildStructuredSections_CardTypesStayInHeroGroup(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Heroes ]
- Wraith
- Card Types:
- Spades: +70% Damage
- Diamond: Cuts enemy resistances by -8% for 5s.`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	if len(heroes.Entries) != 1 {
		t.Fatalf("expected 1 hero entry, got %d", len(heroes.Entries))
	}

	wraith := heroes.Entries[0]
	if wraith.EntityName != "Wraith" {
		t.Fatalf("expected Wraith entry, got %q", wraith.EntityName)
	}
	cardTypes := groupByTitle(wraith, "Card Types")
	if cardTypes == nil {
		t.Fatal("expected Card Types group")
	}
	if len(cardTypes.Changes) != 2 {
		t.Fatalf("expected 2 card type changes, got %d", len(cardTypes.Changes))
	}

	if heroByName(heroes.Entries, "Spades") != nil || heroByName(heroes.Entries, "Diamond") != nil {
		t.Fatal("card type labels should not become standalone hero entries")
	}
}

func TestBuildStructuredSections_ResolvesDoormanAbilitiesWithOrWithoutArticle(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Heroes ]
- The Doorman: Call Bell cooldown increased from 16s to 18s
- Doorman: Hotel Guest cast range increased from 6m to 7m`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}

	doorman := heroByName(heroes.Entries, "Doorman")
	if doorman == nil {
		t.Fatal("expected Doorman entry")
	}
	if groupByTitle(*doorman, "Call Bell") == nil {
		t.Fatal("expected Call Bell group for Doorman")
	}
	if groupByTitle(*doorman, "Hotel Guest") == nil {
		t.Fatal("expected Hotel Guest group for Doorman")
	}
}

func TestBuildStructuredSections_VindcitaAliasDoesNotCreateStandaloneEntry(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Heroes ]
- Vindicta: Bullet damage increased
- Vindcita: Ammo reduced from 22 to 19`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}

	if heroByName(heroes.Entries, "Vindcita") != nil {
		t.Fatal("unexpected standalone Vindcita hero entry")
	}
	if heroByName(heroes.Entries, "Vindicta") == nil {
		t.Fatal("expected Vindicta entry")
	}
}

func sectionByKind(sections []patches.PatchSection, kind string) *patches.PatchSection {
	for i := range sections {
		if sections[i].Kind == kind {
			return &sections[i]
		}
	}
	return nil
}

func heroByName(entries []patches.PatchEntry, name string) *patches.PatchEntry {
	for i := range entries {
		if entries[i].EntityName == name {
			return &entries[i]
		}
	}
	return nil
}

func groupByTitle(entry patches.PatchEntry, title string) *patches.PatchEntryGroup {
	for i := range entry.Groups {
		if entry.Groups[i].Title == title {
			return &entry.Groups[i]
		}
	}
	return nil
}

func testAssetCatalog() *AssetCatalog {
	catalog := &AssetCatalog{
		heroesByNorm: map[string]heroAsset{
			"haze":  {ID: 1, Name: "Haze", Images: heroImages{IconImageSmall: "https://example.test/haze.png"}},
			"bebop": {ID: 2, Name: "Bebop", Images: heroImages{IconImageSmall: "https://example.test/bebop.png"}},
			"calico": {ID: 3, Name: "Calico", Images: heroImages{IconImageSmall: "https://example.test/calico.png"}},
			"wraith": {ID: 4, Name: "Wraith", Images: heroImages{IconImageSmall: "https://example.test/wraith.png"}},
			"the doorman": {ID: 5, Name: "The Doorman", Images: heroImages{IconImageSmall: "https://example.test/doorman.png"}},
			"doorman": {ID: 5, Name: "The Doorman", Images: heroImages{IconImageSmall: "https://example.test/doorman.png"}},
			"vindicta": {ID: 6, Name: "Vindicta", Images: heroImages{IconImageSmall: "https://example.test/vindicta.png"}},
		},
		heroByID: map[int]heroAsset{
			1: {ID: 1, Name: "Haze", Images: heroImages{IconImageSmall: "https://example.test/haze.png"}},
			2: {ID: 2, Name: "Bebop", Images: heroImages{IconImageSmall: "https://example.test/bebop.png"}},
			3: {ID: 3, Name: "Calico", Images: heroImages{IconImageSmall: "https://example.test/calico.png"}},
			4: {ID: 4, Name: "Wraith", Images: heroImages{IconImageSmall: "https://example.test/wraith.png"}},
			5: {ID: 5, Name: "The Doorman", Images: heroImages{IconImageSmall: "https://example.test/doorman.png"}},
			6: {ID: 6, Name: "Vindicta", Images: heroImages{IconImageSmall: "https://example.test/vindicta.png"}},
		},
		itemsByNorm: map[string]itemAsset{
			"active reload": {Name: "Active Reload", Type: "item", ShopImage: "https://example.test/active_reload.png"},
		},
		abilitiesByHero: map[string][]abilityRef{
			"haze": {
				{Name: "Sleep Dagger", NormName: "sleep dagger", Image: "https://example.test/sleep_dagger.png"},
			},
			"bebop": {
				{Name: "Grapple Arm", NormName: "grapple arm", Image: "https://example.test/hook.png"},
				{Name: "Hyper Beam", NormName: "hyper beam", Image: "https://example.test/hyper_beam.png"},
				{Name: "Exploding Uppercut", NormName: "exploding uppercut", Image: "https://example.test/uppercut.png"},
				{Name: "Grapple Arm", NormName: "hook", Image: "https://example.test/hook.png"},
				{Name: "Hyper Beam", NormName: "hyperbeam", Image: "https://example.test/hyper_beam.png"},
				{Name: "Exploding Uppercut", NormName: "uppercut", Image: "https://example.test/uppercut.png"},
			},
			"calico": {
				{Name: "Leaping Slash", NormName: "leaping slash", Image: "https://example.test/leaping_slash.png"},
			},
			"wraith": {
				{Name: "Card Trick", NormName: "card trick", Image: "https://example.test/card_trick.png"},
			},
			"the doorman": {
				{Name: "Call Bell", NormName: "call bell", Image: "https://example.test/call_bell.png"},
				{Name: "Hotel Guest", NormName: "hotel guest", Image: "https://example.test/hotel_guest.png"},
			},
			"vindicta": {
				{Name: "Crow Familiar", NormName: "crow familiar", Image: "https://example.test/crow_familiar.png"},
			},
		},
	}
	return catalog
}
