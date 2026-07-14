package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type storedPatchState struct {
	Slug          string
	Title         string
	Category      string
	Intro         string
	Excerpt       string
	HeroImageURL  string
	PublishedAt   time.Time
	UpdatedAt     time.Time
	SourceType    string
	SourceURL     string
	DetailPayload []byte
}

func newStoredPatchState(
	thread ForumThread,
	detail patchDetailRecord,
	excerpt string,
	publishedAt time.Time,
	updatedAt time.Time,
	detailPayload []byte,
) storedPatchState {
	return storedPatchState{
		Slug:          thread.Slug,
		Title:         detail.Payload.Title,
		Category:      detail.Payload.Category,
		Intro:         detail.Payload.Intro,
		Excerpt:       excerpt,
		HeroImageURL:  detail.Payload.HeroImageURL,
		PublishedAt:   publishedAt,
		UpdatedAt:     updatedAt,
		SourceType:    detail.Payload.Source.Type,
		SourceURL:     detail.Payload.Source.URL,
		DetailPayload: detailPayload,
	}
}

func (stored storedPatchState) matches(desired storedPatchState) bool {
	return stored.Slug == desired.Slug &&
		stored.Title == desired.Title &&
		stored.Category == desired.Category &&
		stored.Intro == desired.Intro &&
		stored.Excerpt == desired.Excerpt &&
		stored.HeroImageURL == desired.HeroImageURL &&
		stored.PublishedAt.Equal(desired.PublishedAt) &&
		stored.UpdatedAt.Equal(desired.UpdatedAt) &&
		stored.SourceType == desired.SourceType &&
		stored.SourceURL == desired.SourceURL &&
		jsonPayloadsEqual(stored.DetailPayload, desired.DetailPayload)
}

func jsonPayloadsEqual(left, right []byte) bool {
	var leftValue any
	var rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func storedTimelineMatches(ctx context.Context, tx *sql.Tx, patchID int64, desired []timelineCandidate) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT block_key, kind, title, source_type, source_url, post_id, released_at, body_text
		FROM patch_release_blocks
		WHERE patch_id = $1
		ORDER BY sort_order
	`, patchID)
	if err != nil {
		return false, fmt.Errorf("load stored timeline: %w", err)
	}
	defer rows.Close()

	index := 0
	for rows.Next() {
		if index >= len(desired) {
			return false, nil
		}
		stored, err := scanStoredTimelineBlock(rows)
		if err != nil {
			return false, err
		}
		if !timelineCandidatesEqual(stored, desired[index]) {
			return false, nil
		}
		index++
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate stored timeline: %w", err)
	}
	return index == len(desired), nil
}

func timelineCandidatesEqual(left, right timelineCandidate) bool {
	return left.Key == right.Key &&
		left.Kind == right.Kind &&
		left.Title == right.Title &&
		left.SourceType == right.SourceType &&
		left.SourceURL == right.SourceURL &&
		left.PostID == right.PostID &&
		left.ReleasedAt.Equal(right.ReleasedAt) &&
		left.BodyText == right.BodyText
}
