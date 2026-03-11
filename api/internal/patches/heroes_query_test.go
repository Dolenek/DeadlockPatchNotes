package patches

import (
	"testing"
	"time"
)

func TestBuildHeroList_ReturnsHeroesWithChanges(t *testing.T) {
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
									ID:         "abrams",
									EntityName: "Abrams",
									Changes:    []PatchChange{{ID: "c1", Text: "Base health increased"}},
								},
							},
						},
					},
				},
			},
		},
	}

	payload := buildHeroList(details)
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 hero, got %d", len(payload.Items))
	}
	if payload.Items[0].Slug != "abrams" {
		t.Fatalf("expected abrams slug, got %q", payload.Items[0].Slug)
	}
}

func TestBuildHeroChanges_AppliesSkillAndDateFilters(t *testing.T) {
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
									ID:         "abrams",
									EntityName: "Abrams",
									Groups: []PatchEntryGroup{
										{
											ID:      "shoulder-charge",
											Title:   "Shoulder Charge",
											Changes: []PatchChange{{ID: "c1", Text: "Cooldown reduced"}},
										},
									},
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
							ID:    "heroes",
							Title: "Heroes",
							Kind:  "heroes",
							Entries: []PatchEntry{
								{
									ID:         "abrams",
									EntityName: "Abrams",
									Groups: []PatchEntryGroup{
										{
											ID:      "seismic-impact",
											Title:   "Seismic Impact",
											Changes: []PatchChange{{ID: "c2", Text: "Duration increased"}},
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

	from := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.March, 8, 23, 59, 59, 0, time.UTC)
	payload, err := buildHeroChanges(details, HeroChangesQuery{
		HeroSlug: "abrams",
		Skill:    "Shoulder Charge",
		From:     &from,
		To:       &to,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 filtered timeline block, got %d", len(payload.Items))
	}
	if len(payload.Items[0].Skills) != 1 || payload.Items[0].Skills[0].Title != "Shoulder Charge" {
		t.Fatalf("unexpected skills payload: %+v", payload.Items[0].Skills)
	}
}
