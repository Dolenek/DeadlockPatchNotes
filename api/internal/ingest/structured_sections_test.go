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

func TestBuildStructuredSections_HeroAbilityPrefixesBeatItemResolution(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Heroes ]
- Abrams
- Siphon Life: Added new animation logic support for Siphon Life and items.
- Shoulder Charge: Base duration increased from 1.2s to 1.4s.
- Shoulder Charge: T1 is now "On Hero Collide: +25% Weapon Damage for 8s".`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	if items := sectionByKind(sections, "items"); items != nil && len(items.Entries) != 0 {
		t.Fatalf("expected no item entries, got %+v", items.Entries)
	}

	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	abrams := heroByName(heroes.Entries, "Abrams")
	if abrams == nil {
		t.Fatal("expected Abrams entry")
	}
	if groupByTitle(*abrams, "Siphon Life") == nil {
		t.Fatal("expected Siphon Life group under Abrams")
	}
	shoulderCharge := groupByTitle(*abrams, "Shoulder Charge")
	if shoulderCharge == nil {
		t.Fatal("expected Shoulder Charge group under Abrams")
	}
	if len(shoulderCharge.Changes) != 2 {
		t.Fatalf("expected 2 Shoulder Charge changes, got %+v", shoulderCharge.Changes)
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

func TestBuildStructuredSections_KeepsDoormanFollowupAbilityLinesOutOfGeneral(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Heroes ]
- Doorman
- Gun now pierces through targets at 50% reduced damage
- Call Bell time between charges increased from 4s to 6s
- Doorway now has a timer icon above the ability
- Luggage Cart is now 20% larger (20% wider hitbox as well)
- Hotel Guest cast range increased from 6m to 7m`,
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
	if len(doorman.Changes) != 1 {
		t.Fatalf("expected 1 Doorman general change, got %d", len(doorman.Changes))
	}
	if doorman.Changes[0].Text != "Gun now pierces through targets at 50% reduced damage" {
		t.Fatalf("unexpected Doorman general change: %+v", doorman.Changes)
	}

	for _, title := range []string{"Call Bell", "Doorway", "Luggage Cart", "Hotel Guest"} {
		if groupByTitle(*doorman, title) == nil {
			t.Fatalf("expected %s group for Doorman", title)
		}
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

func TestBuildStructuredSections_RepairsAfflictionOutOfItems(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Items ]
- Active Reload: Reload speed increased
- Affliction: Duration reduced from 18s to 14s`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	items := sectionByKind(sections, "items")
	if items == nil {
		t.Fatal("expected items section")
	}
	if len(items.Entries) != 1 || items.Entries[0].EntityName != "Active Reload" {
		t.Fatalf("expected only Active Reload to remain in items, got %+v", items.Entries)
	}

	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	pocket := heroByName(heroes.Entries, "Pocket")
	if pocket == nil {
		t.Fatal("expected Pocket entry")
	}
	affliction := groupByTitle(*pocket, "Affliction")
	if affliction == nil {
		t.Fatal("expected Affliction group under Pocket")
	}
	if len(affliction.Changes) != 1 || affliction.Changes[0].Text != "Duration reduced from 18s to 14s" {
		t.Fatalf("unexpected Affliction changes: %+v", affliction.Changes)
	}
}

func TestBuildStructuredSections_RepairsCrimsonSlashOutOfItems(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Items ]
- Crimson Slash: adjusted slash effects height to be better aligned to crosshair.`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	if items := sectionByKind(sections, "items"); items != nil && len(items.Entries) != 0 {
		t.Fatalf("expected Crimson Slash to leave items, got %+v", items.Entries)
	}

	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	yamato := heroByName(heroes.Entries, "Yamato")
	if yamato == nil {
		t.Fatal("expected Yamato entry")
	}
	crimsonSlash := groupByTitle(*yamato, "Crimson Slash")
	if crimsonSlash == nil {
		t.Fatal("expected Crimson Slash group under Yamato")
	}
	if len(crimsonSlash.Changes) != 1 || crimsonSlash.Changes[0].Text != "adjusted slash effects height to be better aligned to crosshair." {
		t.Fatalf("unexpected Crimson Slash changes: %+v", crimsonSlash.Changes)
	}
}

func TestBuildStructuredSections_LeavesAmbiguousAbilityNamesInItems(t *testing.T) {
	catalog := testAssetCatalog()
	catalog.abilitiesByNorm["shared ability"] = []abilityOwnerRef{
		{
			HeroKey:             "abrams",
			HeroName:            "Abrams",
			HeroIconFallbackURL: "https://example.test/abrams.png",
			AbilityName:         "Shared Ability",
			AbilityNormName:     "shared ability",
			AbilityIconFallbackURL: "https://example.test/shared-1.png",
		},
		{
			HeroKey:             "bebop",
			HeroName:            "Bebop",
			HeroIconFallbackURL: "https://example.test/bebop.png",
			AbilityName:         "Shared Ability",
			AbilityNormName:     "shared ability",
			AbilityIconFallbackURL: "https://example.test/shared-2.png",
		},
	}

	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Items ]
- Shared Ability: Cooldown reduced`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)
	items := sectionByKind(sections, "items")
	if items == nil || len(items.Entries) != 1 || items.Entries[0].EntityName != "Shared Ability" {
		t.Fatalf("expected ambiguous Shared Ability to remain in items, got %+v", items)
	}
	if heroes := sectionByKind(sections, "heroes"); heroes != nil && len(heroes.Entries) != 0 {
		t.Fatalf("expected no hero entries for ambiguous ability, got %+v", heroes.Entries)
	}
}

func TestBuildStructuredSections_ParsesExtendedSteamSectionHeaders(t *testing.T) {
	catalog := testAssetCatalog()
	blocks := []timelineCandidate{
		{
			Key: "post-1-steam-1",
			BodyText: `[ Hero Content Improvements ]
- Yamato
- Crimson Slash: adjusted slash effects height to be better aligned to crosshair.
[ Item Content Improvements ]
- Active Reload
- Now plays a sound when entering the active window.
- Magic Carpet
- Fixed a bug that caused the ambient sound to not loop properly.`,
		},
	}

	sections := buildStructuredSections(blocks, catalog)

	heroes := sectionByKind(sections, "heroes")
	if heroes == nil {
		t.Fatal("expected heroes section")
	}
	yamato := heroByName(heroes.Entries, "Yamato")
	if yamato == nil {
		t.Fatal("expected Yamato entry")
	}
	if groupByTitle(*yamato, "Crimson Slash") == nil {
		t.Fatal("expected Crimson Slash group under Yamato")
	}

	items := sectionByKind(sections, "items")
	if items == nil {
		t.Fatal("expected items section")
	}
	activeReload := itemEntryByName(items.Entries, "Active Reload")
	if activeReload == nil || len(activeReload.Changes) != 1 {
		t.Fatalf("expected Active Reload item entry with one change, got %+v", activeReload)
	}
	magicCarpet := itemEntryByName(items.Entries, "Magic Carpet")
	if magicCarpet == nil || len(magicCarpet.Changes) != 1 {
		t.Fatalf("expected Magic Carpet item entry with one change, got %+v", magicCarpet)
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

func itemEntryByName(entries []patches.PatchEntry, name string) *patches.PatchEntry {
	for i := range entries {
		if entries[i].EntityName == name {
			return &entries[i]
		}
	}
	return nil
}

func testAssetCatalog() *AssetCatalog {
	catalog := &AssetCatalog{
		heroesByNorm: map[string]heroAsset{
			"haze":  {ID: 1, Name: "Haze", Images: heroImages{IconImageSmall: "https://example.test/haze.png"}},
			"abrams": {ID: 9, Name: "Abrams", Images: heroImages{IconImageSmall: "https://example.test/abrams.png"}},
			"pocket": {ID: 7, Name: "Pocket", Images: heroImages{IconImageSmall: "https://example.test/pocket.png"}},
			"bebop": {ID: 2, Name: "Bebop", Images: heroImages{IconImageSmall: "https://example.test/bebop.png"}},
			"calico": {ID: 3, Name: "Calico", Images: heroImages{IconImageSmall: "https://example.test/calico.png"}},
			"wraith": {ID: 4, Name: "Wraith", Images: heroImages{IconImageSmall: "https://example.test/wraith.png"}},
			"the doorman": {ID: 5, Name: "The Doorman", Images: heroImages{IconImageSmall: "https://example.test/doorman.png"}},
			"doorman": {ID: 5, Name: "The Doorman", Images: heroImages{IconImageSmall: "https://example.test/doorman.png"}},
			"vindicta": {ID: 6, Name: "Vindicta", Images: heroImages{IconImageSmall: "https://example.test/vindicta.png"}},
			"yamato": {ID: 8, Name: "Yamato", Images: heroImages{IconImageSmall: "https://example.test/yamato.png"}},
		},
		heroByID: map[int]heroAsset{
			1: {ID: 1, Name: "Haze", Images: heroImages{IconImageSmall: "https://example.test/haze.png"}},
			2: {ID: 2, Name: "Bebop", Images: heroImages{IconImageSmall: "https://example.test/bebop.png"}},
			3: {ID: 3, Name: "Calico", Images: heroImages{IconImageSmall: "https://example.test/calico.png"}},
			4: {ID: 4, Name: "Wraith", Images: heroImages{IconImageSmall: "https://example.test/wraith.png"}},
			5: {ID: 5, Name: "The Doorman", Images: heroImages{IconImageSmall: "https://example.test/doorman.png"}},
			6: {ID: 6, Name: "Vindicta", Images: heroImages{IconImageSmall: "https://example.test/vindicta.png"}},
			7: {ID: 7, Name: "Pocket", Images: heroImages{IconImageSmall: "https://example.test/pocket.png"}},
			8: {ID: 8, Name: "Yamato", Images: heroImages{IconImageSmall: "https://example.test/yamato.png"}},
			9: {ID: 9, Name: "Abrams", Images: heroImages{IconImageSmall: "https://example.test/abrams.png"}},
		},
		itemsByNorm: map[string]itemAsset{
			"active reload":   {Name: "Active Reload", Type: "item", ShopImage: "https://example.test/active_reload.png"},
			"magic carpet":    {Name: "Magic Carpet", Type: "item", ShopImage: "https://example.test/magic_carpet.png"},
			"affliction":      {Name: "Affliction", Type: "ability", Image: "https://example.test/affliction.png", HeroID: 7},
			"crimson slash":   {Name: "Crimson Slash", Type: "ability", Image: "https://example.test/crimson_slash.png", HeroID: 8},
			"siphon life":     {Name: "Siphon Life", Type: "ability", Image: "https://example.test/siphon_life.png", HeroID: 9},
			"shoulder charge": {Name: "Shoulder Charge", Type: "ability", Image: "https://example.test/shoulder_charge.png", HeroID: 9},
		},
		abilitiesByHero: map[string][]abilityRef{
			"haze": {
				{Name: "Sleep Dagger", NormName: "sleep dagger", Image: "https://example.test/sleep_dagger.png"},
			},
			"abrams": {
				{Name: "Siphon Life", NormName: "siphon life", Image: "https://example.test/siphon_life.png"},
				{Name: "Shoulder Charge", NormName: "shoulder charge", Image: "https://example.test/shoulder_charge.png"},
			},
			"pocket": {
				{Name: "Affliction", NormName: "affliction", Image: "https://example.test/affliction.png"},
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
			"doorman": {
				{Name: "Call Bell", NormName: "call bell", Image: "https://example.test/call_bell.png"},
				{Name: "Doorway", NormName: "doorway", Image: "https://example.test/doorway.png"},
				{Name: "Luggage Cart", NormName: "luggage cart", Image: "https://example.test/luggage_cart.png"},
				{Name: "Hotel Guest", NormName: "hotel guest", Image: "https://example.test/hotel_guest.png"},
			},
			"vindicta": {
				{Name: "Crow Familiar", NormName: "crow familiar", Image: "https://example.test/crow_familiar.png"},
			},
			"yamato": {
				{Name: "Crimson Slash", NormName: "crimson slash", Image: "https://example.test/crimson_slash.png"},
			},
		},
		abilitiesByNorm: map[string][]abilityOwnerRef{
			"sleep dagger": {
				{HeroKey: "haze", HeroName: "Haze", HeroIconFallbackURL: "https://example.test/haze.png", AbilityName: "Sleep Dagger", AbilityNormName: "sleep dagger", AbilityIconFallbackURL: "https://example.test/sleep_dagger.png"},
			},
			"siphon life": {
				{HeroKey: "abrams", HeroName: "Abrams", HeroIconFallbackURL: "https://example.test/abrams.png", AbilityName: "Siphon Life", AbilityNormName: "siphon life", AbilityIconFallbackURL: "https://example.test/siphon_life.png"},
			},
			"shoulder charge": {
				{HeroKey: "abrams", HeroName: "Abrams", HeroIconFallbackURL: "https://example.test/abrams.png", AbilityName: "Shoulder Charge", AbilityNormName: "shoulder charge", AbilityIconFallbackURL: "https://example.test/shoulder_charge.png"},
			},
			"affliction": {
				{HeroKey: "pocket", HeroName: "Pocket", HeroIconFallbackURL: "https://example.test/pocket.png", AbilityName: "Affliction", AbilityNormName: "affliction", AbilityIconFallbackURL: "https://example.test/affliction.png"},
			},
			"grapple arm": {
				{HeroKey: "bebop", HeroName: "Bebop", HeroIconFallbackURL: "https://example.test/bebop.png", AbilityName: "Grapple Arm", AbilityNormName: "grapple arm", AbilityIconFallbackURL: "https://example.test/hook.png"},
			},
			"hook": {
				{HeroKey: "bebop", HeroName: "Bebop", HeroIconFallbackURL: "https://example.test/bebop.png", AbilityName: "Grapple Arm", AbilityNormName: "hook", AbilityIconFallbackURL: "https://example.test/hook.png"},
			},
			"hyper beam": {
				{HeroKey: "bebop", HeroName: "Bebop", HeroIconFallbackURL: "https://example.test/bebop.png", AbilityName: "Hyper Beam", AbilityNormName: "hyper beam", AbilityIconFallbackURL: "https://example.test/hyper_beam.png"},
			},
			"hyperbeam": {
				{HeroKey: "bebop", HeroName: "Bebop", HeroIconFallbackURL: "https://example.test/bebop.png", AbilityName: "Hyper Beam", AbilityNormName: "hyperbeam", AbilityIconFallbackURL: "https://example.test/hyper_beam.png"},
			},
			"exploding uppercut": {
				{HeroKey: "bebop", HeroName: "Bebop", HeroIconFallbackURL: "https://example.test/bebop.png", AbilityName: "Exploding Uppercut", AbilityNormName: "exploding uppercut", AbilityIconFallbackURL: "https://example.test/uppercut.png"},
			},
			"uppercut": {
				{HeroKey: "bebop", HeroName: "Bebop", HeroIconFallbackURL: "https://example.test/bebop.png", AbilityName: "Exploding Uppercut", AbilityNormName: "uppercut", AbilityIconFallbackURL: "https://example.test/uppercut.png"},
			},
			"leaping slash": {
				{HeroKey: "calico", HeroName: "Calico", HeroIconFallbackURL: "https://example.test/calico.png", AbilityName: "Leaping Slash", AbilityNormName: "leaping slash", AbilityIconFallbackURL: "https://example.test/leaping_slash.png"},
			},
			"card trick": {
				{HeroKey: "wraith", HeroName: "Wraith", HeroIconFallbackURL: "https://example.test/wraith.png", AbilityName: "Card Trick", AbilityNormName: "card trick", AbilityIconFallbackURL: "https://example.test/card_trick.png"},
			},
			"call bell": {
				{HeroKey: "doorman", HeroName: "Doorman", HeroIconFallbackURL: "https://example.test/doorman.png", AbilityName: "Call Bell", AbilityNormName: "call bell", AbilityIconFallbackURL: "https://example.test/call_bell.png"},
			},
			"doorway": {
				{HeroKey: "doorman", HeroName: "Doorman", HeroIconFallbackURL: "https://example.test/doorman.png", AbilityName: "Doorway", AbilityNormName: "doorway", AbilityIconFallbackURL: "https://example.test/doorway.png"},
			},
			"luggage cart": {
				{HeroKey: "doorman", HeroName: "Doorman", HeroIconFallbackURL: "https://example.test/doorman.png", AbilityName: "Luggage Cart", AbilityNormName: "luggage cart", AbilityIconFallbackURL: "https://example.test/luggage_cart.png"},
			},
			"hotel guest": {
				{HeroKey: "doorman", HeroName: "Doorman", HeroIconFallbackURL: "https://example.test/doorman.png", AbilityName: "Hotel Guest", AbilityNormName: "hotel guest", AbilityIconFallbackURL: "https://example.test/hotel_guest.png"},
			},
			"crow familiar": {
				{HeroKey: "vindicta", HeroName: "Vindicta", HeroIconFallbackURL: "https://example.test/vindicta.png", AbilityName: "Crow Familiar", AbilityNormName: "crow familiar", AbilityIconFallbackURL: "https://example.test/crow_familiar.png"},
			},
			"crimson slash": {
				{HeroKey: "yamato", HeroName: "Yamato", HeroIconFallbackURL: "https://example.test/yamato.png", AbilityName: "Crimson Slash", AbilityNormName: "crimson slash", AbilityIconFallbackURL: "https://example.test/crimson_slash.png"},
			},
		},
	}
	return catalog
}
