package ingest

import "testing"

func TestTimelineKindForPost(t *testing.T) {
	tests := []struct {
		name         string
		firstPost    bool
		blockIndex   int
		parsedKind   string
		expectedKind string
	}{
		{name: "first block of first post", firstPost: true, blockIndex: 0, parsedKind: "initial", expectedKind: "initial"},
		{name: "first block of later post", firstPost: false, blockIndex: 0, parsedKind: "initial", expectedKind: "hotfix"},
		{name: "embedded hotfix", firstPost: false, blockIndex: 1, parsedKind: "hotfix", expectedKind: "hotfix"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := timelineKindForPost(test.firstPost, test.blockIndex, test.parsedKind); got != test.expectedKind {
				t.Fatalf("expected %q, got %q", test.expectedKind, got)
			}
		})
	}
}
