package patches

import (
	"encoding/json"
	"testing"
)

func TestPatchDetailUnmarshal_LegacyKeys(t *testing.T) {
	raw := []byte(`{
		"id":"patch-1",
		"slug":"03-06-2026-update",
		"title":"03-06-2026 Update",
		"publishedAt":"2026-03-06T22:36:00Z",
		"category":"Regular Update",
		"source":{"type":"forum","url":"https://example.test"},
		"heroImageUrl":"https://example.test/hero.png",
		"intro":"Example",
		"sections":[{"id":"general","title":"General","kind":"general","entries":[{"id":"e1","entityName":"Core Gameplay","changes":[{"id":"c1","text":"Line"}]}]}],
		"timeline":[
			{"id":"initial","kind":"initial","title":"Initial Update","releasedAt":"2026-03-06T22:36:00Z","source":{"type":"forum","url":"https://example.test"},"changes":[{"id":"c1","text":"Initial"}]},
			{"id":"hotfix-1","kind":"hotfix","title":"Hotfix 2026-03-07","releasedAt":"2026-03-07T18:11:00Z","source":{"type":"forum","url":"https://example.test"},"changes":[{"id":"c2","text":"Hotfix"}]}
		]
	}`)

	var detail PatchDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		t.Fatalf("decode legacy patch detail: %v", err)
	}

	if detail.HeroImageURL != "https://example.test/hero.png" {
		t.Fatalf("expected legacy heroImageUrl to map to imageUrl field, got %q", detail.HeroImageURL)
	}
	if len(detail.Timeline) != 2 {
		t.Fatalf("expected 2 timeline blocks from legacy timeline key, got %d", len(detail.Timeline))
	}
	if detail.Timeline[0].Kind != "initial" || detail.Timeline[1].Kind != "hotfix" {
		t.Fatalf("expected legacy kind values to map, got %+v", detail.Timeline)
	}
}

func TestPatchDetailUnmarshal_NewKeys(t *testing.T) {
	raw := []byte(`{
		"id":"patch-1",
		"slug":"03-06-2026-update",
		"title":"03-06-2026 Update",
		"publishedAt":"2026-03-06T22:36:00Z",
		"category":"Regular Update",
		"source":{"type":"forum","url":"https://example.test"},
		"imageUrl":"https://example.test/hero.png",
		"intro":"Example",
		"sections":[{"id":"general","title":"General","kind":"general","entries":[{"id":"e1","entityName":"Core Gameplay","changes":[{"id":"c1","text":"Line"}]}]}],
		"releaseTimeline":[
			{"id":"initial","releaseType":"initial","title":"Initial Update","releasedAt":"2026-03-06T22:36:00Z","source":{"type":"forum","url":"https://example.test"},"changes":[{"id":"c1","text":"Initial"}]}
		]
	}`)

	var detail PatchDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		t.Fatalf("decode new patch detail: %v", err)
	}

	if detail.HeroImageURL != "https://example.test/hero.png" {
		t.Fatalf("expected imageUrl to decode, got %q", detail.HeroImageURL)
	}
	if len(detail.Timeline) != 1 {
		t.Fatalf("expected 1 timeline block from releaseTimeline key, got %d", len(detail.Timeline))
	}
	if detail.Timeline[0].Kind != "initial" {
		t.Fatalf("expected releaseType to decode, got %q", detail.Timeline[0].Kind)
	}
}
