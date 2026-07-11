package patches

import "time"

type timelineEventOwner struct {
	detailIndex int
	blockIndex  int
	distance    time.Duration
}

// deduplicateTimelineEvents removes repeated source events from aggregate
// histories without changing the patch-detail payload retained in detailBySlug.
func deduplicateTimelineEvents(details []PatchDetail) []PatchDetail {
	owners := make(map[string]timelineEventOwner)
	for detailIndex, detail := range details {
		publishedAt := parseRFC3339(detail.PublishedAt)
		for blockIndex, block := range detail.Timeline {
			signature := blockBodySignature(block)
			if signature == "" {
				continue
			}
			candidate := timelineEventOwner{
				detailIndex: detailIndex,
				blockIndex:  blockIndex,
				distance:    timestampDistance(publishedAt, parseRFC3339(block.ReleasedAt)),
			}
			current, found := owners[signature]
			currentPublishedAt := parseRFC3339(details[current.detailIndex].PublishedAt)
			if !found || candidate.distance < current.distance ||
				(candidate.distance == current.distance && publishedAt.After(currentPublishedAt)) {
				owners[signature] = candidate
			}
		}
	}

	result := make([]PatchDetail, len(details))
	copy(result, details)
	for detailIndex := range result {
		filtered := make([]PatchTimelineBlock, 0, len(result[detailIndex].Timeline))
		for blockIndex, block := range result[detailIndex].Timeline {
			signature := blockBodySignature(block)
			owner, found := owners[signature]
			if signature != "" && found && (owner.detailIndex != detailIndex || owner.blockIndex != blockIndex) {
				continue
			}
			filtered = append(filtered, block)
		}
		result[detailIndex].Timeline = filtered
	}
	return result
}

func timestampDistance(left, right time.Time) time.Duration {
	if left.IsZero() || right.IsZero() {
		return time.Duration(1<<63 - 1)
	}
	distance := left.Sub(right)
	if distance < 0 {
		return -distance
	}
	return distance
}
