package ingest

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"deadlockpatchnotes/api/internal/patches"
)

type patchDetailRecord struct {
	Payload patches.PatchDetail
	Excerpt string
}

func buildPatchFromThread(ctx context.Context, client *http.Client, thread ForumThread, catalog *AssetCatalog) (patchDetailRecord, []timelineCandidate, time.Time, time.Time) {
	blocks := make([]timelineCandidate, 0, len(thread.Posts)+2)
	coverImage := ""

	for index, post := range thread.Posts {
		isFirstPost := index == 0
		hasSteamContent := false
		if post.SteamURL != "" {
			event, err := FetchSteamEvent(ctx, client, post.SteamURL, post.PublishedAt)
			if err == nil {
				hasSteamContent = len(event.BodyBlocks) > 0
				if coverImage == "" {
					coverImage = firstNonEmpty(event.HeroImage, post.SteamImage)
				}
				for blockIndex, block := range event.BodyBlocks {
					kind := block.Kind
					if isFirstPost && blockIndex == 0 {
						kind = "initial"
					}
					blocks = append(blocks, timelineCandidate{
						Key:        fmt.Sprintf("post-%s-steam-%d", post.PostID, blockIndex+1),
						Kind:       kind,
						Title:      firstNonEmpty(block.Title, event.Title),
						SourceType: "steam-news",
						SourceURL:  post.SteamURL,
						PostID:     post.PostID,
						ReleasedAt: nonZeroTime(block.ReleasedAt, post.PublishedAt),
						BodyText:   strings.TrimSpace(block.BodyText),
					})
				}
			}
		}

		if hasSteamContent {
			continue
		}

		if strings.TrimSpace(post.BodyText) != "" {
			kind := "hotfix"
			title := fmt.Sprintf("Hotfix %s", post.PublishedAt.UTC().Format("2006-01-02"))
			if isFirstPost && len(blocks) == 0 {
				kind = "initial"
				title = "Initial Update"
			}
			blocks = append(blocks, timelineCandidate{
				Key:        fmt.Sprintf("post-%s-forum", post.PostID),
				Kind:       kind,
				Title:      title,
				SourceType: "forum-post",
				SourceURL:  post.ForumPostURL,
				PostID:     post.PostID,
				ReleasedAt: post.PublishedAt,
				BodyText:   strings.TrimSpace(post.BodyText),
			})
		}
	}

	blocks = dedupeBlocks(blocks)
	sort.SliceStable(blocks, func(i, j int) bool {
		if blocks[i].ReleasedAt.Equal(blocks[j].ReleasedAt) {
			return blocks[i].Key < blocks[j].Key
		}
		return blocks[i].ReleasedAt.Before(blocks[j].ReleasedAt)
	})

	if len(blocks) > 0 && blocks[0].Kind != "initial" {
		blocks[0].Kind = "initial"
		blocks[0].Title = "Initial Update"
	}

	if len(blocks) == 0 {
		return patchDetailRecord{}, nil, time.Time{}, time.Time{}
	}

	publishedAt := blocks[0].ReleasedAt
	updatedAt := blocks[len(blocks)-1].ReleasedAt

	payload := buildDetailPayload(thread, blocks, coverImage, catalog)
	excerpt := buildIntro(payload.Sections[0].Entries)

	return patchDetailRecord{Payload: payload, Excerpt: excerpt}, blocks, publishedAt, updatedAt
}

func buildDetailPayload(thread ForumThread, blocks []timelineCandidate, coverImage string, catalog *AssetCatalog) patches.PatchDetail {
	timeline := make([]patches.PatchTimelineBlock, 0, len(blocks))

	for _, block := range blocks {
		blockSections := buildStructuredSections([]timelineCandidate{block}, catalog)
		if len(blockSections) == 0 {
			blockSections = fallbackSectionsFromTimeline([]timelineCandidate{block})
		}
		changeLines := toChangeLines(block.BodyText, block.Key)
		timeline = append(timeline, patches.PatchTimelineBlock{
			ID:         block.Key,
			Kind:       block.Kind,
			Title:      block.Title,
			ReleasedAt: block.ReleasedAt.UTC().Format(time.RFC3339),
			Source: patches.PatchSource{
				Type: block.SourceType,
				URL:  block.SourceURL,
			},
			Changes:  changeLines,
			Sections: blockSections,
		})
	}

	sections := buildStructuredSections(blocks, catalog)
	if len(sections) == 0 {
		sections = fallbackSectionsFromTimeline(blocks)
	}

	intro := buildIntro(sections[0].Entries)
	source := patches.PatchSource{Type: blocks[0].SourceType, URL: blocks[0].SourceURL}

	return patches.PatchDetail{
		ID:           fmt.Sprintf("%d", thread.ThreadID),
		Slug:         thread.Slug,
		Title:        thread.Title,
		PublishedAt:  blocks[0].ReleasedAt.UTC().Format(time.RFC3339),
		Category:     "Regular Update",
		Source:       source,
		HeroImageURL: coverImage,
		Intro:        intro,
		Sections:     sections,
		Timeline: timeline,
	}
}

func fallbackSectionsFromTimeline(blocks []timelineCandidate) []patches.PatchSection {
	entries := make([]patches.PatchEntry, 0, len(blocks))
	for index, block := range blocks {
		entryName := block.Title
		if entryName == "" {
			if block.Kind == "initial" {
				entryName = "Initial Update"
			} else {
				entryName = fmt.Sprintf("Hotfix %d", index)
			}
		}
		entries = append(entries, patches.PatchEntry{
			ID:         block.Key,
			EntityName: entryName,
			Changes:    toChangeLines(block.BodyText, block.Key),
		})
	}

	if len(entries) == 0 {
		entries = append(entries, patches.PatchEntry{
			ID:         "general-gameplay",
			EntityName: "Core Gameplay",
			Changes: []patches.PatchChange{
				{ID: "general-1", Text: "No line-item changes listed."},
			},
		})
	}

	return []patches.PatchSection{
		{
			ID:      "general",
			Title:   "General",
			Kind:    "general",
			Entries: entries,
		},
	}
}

func buildIntro(entries []patches.PatchEntry) string {
	if len(entries) == 0 {
		return "Deadlock patch update."
	}
	for _, change := range entries[0].Changes {
		text := strings.TrimSpace(change.Text)
		if text == "" {
			continue
		}
		if len(text) <= 220 {
			return text
		}
		return strings.TrimSpace(text[:217]) + "..."
	}
	return "Deadlock patch update."
}

func toChangeLines(body, prefix string) []patches.PatchChange {
	lines := strings.Split(body, "\n")
	changes := make([]patches.PatchChange, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		changes = append(changes, patches.PatchChange{
			ID:   fmt.Sprintf("%s-%d", prefix, len(changes)+1),
			Text: line,
		})
	}
	if len(changes) == 0 {
		changes = append(changes, patches.PatchChange{ID: prefix + "-1", Text: "No line-item changes listed."})
	}
	return changes
}

func dedupeBlocks(input []timelineCandidate) []timelineCandidate {
	seen := map[string]bool{}
	output := make([]timelineCandidate, 0, len(input))
	for _, block := range input {
		normalizedBody := normalizeBodyForHash(block.BodyText)
		if normalizedBody == "" {
			continue
		}
		hash := hashText(normalizedBody)
		if seen[hash] {
			continue
		}
		seen[hash] = true
		block.BodyText = normalizedBody
		output = append(output, block)
	}
	return output
}

func normalizeBodyForHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lines := strings.Split(value, "\n")
	normalized := make([]string, 0, len(lines))
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.TrimSpace(strings.Join(normalized, "\n"))
}

func hashText(value string) string {
	sum := sha1.Sum([]byte(strings.ToLower(value)))
	return hex.EncodeToString(sum[:])
}

func nonZeroTime(candidate, fallback time.Time) time.Time {
	if !candidate.IsZero() {
		return candidate
	}
	return fallback
}
