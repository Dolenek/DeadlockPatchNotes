package ingest

import (
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchSteamEventParsesMetadataImagesAndHotfixes(t *testing.T) {
	published := time.Date(2026, time.July, 1, 8, 30, 0, 0, time.UTC)
	envelope := steamEventEnvelope{
		GID: "1", EventName: "Fallback title",
		AnnouncementBody: steamAnnouncement{
			Headline: "Patch title", PostTime: published.Unix(),
			Body: `[p]- Initial change[/p][h3]07-02-2026 Patch:[/h3][p]- Hotfix change[/p]`,
		},
	}
	server := steamEventServer(t, []steamEventEnvelope{envelope}, `<meta property="og:image" content="https://images.test/hero.png?a=1&amp;b=2">`)
	defer server.Close()

	event, err := FetchSteamEvent(t.Context(), server.Client(), server.URL, published.Add(-time.Hour))
	if err != nil {
		t.Fatalf("fetch Steam event: %v", err)
	}
	if event.Title != "Patch title" || !event.Published.Equal(published) {
		t.Fatalf("unexpected metadata: %+v", event)
	}
	if event.HeroImage != "https://images.test/hero.png?a=1&b=2" || len(event.BodyBlocks) != 2 {
		t.Fatalf("unexpected image or blocks: image=%q blocks=%+v", event.HeroImage, event.BodyBlocks)
	}
	assertSteamBlock(t, event.BodyBlocks[0], "initial", "- Initial change", published)
	expectedHotfix := time.Date(2026, time.July, 2, 12, 0, 0, 0, time.UTC)
	assertSteamBlock(t, event.BodyBlocks[1], "hotfix", "- Hotfix change", expectedHotfix)
}

func TestFetchSteamEventUsesFallbackTimeAndCapsuleImage(t *testing.T) {
	fallback := time.Date(2026, time.June, 30, 9, 0, 0, 0, time.UTC)
	capsule, _ := json.Marshal(steamJSONData{LocalizedCapsuleImage: []*string{nil, stringPointer("capsule.png")}})
	envelope := steamEventEnvelope{
		EventName: "Event title", JSONData: string(capsule),
		AnnouncementBody: steamAnnouncement{Body: `[p]- Change[/p]`},
	}
	server := steamEventServer(t, []steamEventEnvelope{envelope}, "")
	defer server.Close()

	event, err := FetchSteamEvent(t.Context(), server.Client(), server.URL, fallback)
	if err != nil {
		t.Fatalf("fetch Steam event: %v", err)
	}
	if event.Title != "Event title" || !event.Published.Equal(fallback) {
		t.Fatalf("fallback metadata was not used: %+v", event)
	}
	if !strings.HasSuffix(event.HeroImage, "/capsule.png") {
		t.Fatalf("capsule image was not used: %q", event.HeroImage)
	}
}

func TestFetchSteamEventRejectsInvalidMetadata(t *testing.T) {
	for _, testCase := range []struct {
		name, body, errorPart string
	}{
		{name: "missing", body: "<html></html>", errorPart: "metadata missing"},
		{name: "invalid JSON", body: `<div data-partnereventstore="{oops}"></div>`, errorPart: "decode steam event payload"},
		{name: "empty", body: `<div data-partnereventstore="[]"></div>`, errorPart: "empty steam event payload"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(testCase.body)) }))
			defer server.Close()
			_, err := FetchSteamEvent(t.Context(), server.Client(), server.URL, time.Time{})
			if err == nil || !strings.Contains(err.Error(), testCase.errorPart) {
				t.Fatalf("expected %q error, got %v", testCase.errorPart, err)
			}
		})
	}
}

func TestSplitSteamBodyBlocksUsesFallbackForInvalidPatchDate(t *testing.T) {
	fallback := time.Date(2026, time.July, 3, 5, 0, 0, 0, time.UTC)
	blocks := splitSteamBodyBlocks("Initial\n13-40-2026 Patch:\nHotfix", fallback)
	if len(blocks) != 2 || !blocks[1].ReleasedAt.Equal(fallback) || blocks[1].Title != "Hotfix 2026-07-03" {
		t.Fatalf("unexpected invalid-date fallback: %+v", blocks)
	}
}

func steamEventServer(t *testing.T, events []steamEventEnvelope, head string) *httptest.Server {
	t.Helper()
	payload, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal Steam fixture: %v", err)
	}
	body := head + `<div data-partnereventstore="` + html.EscapeString(string(payload)) + `"></div>`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }))
}

func assertSteamBlock(t *testing.T, block SteamBodyBlock, kind, body string, releasedAt time.Time) {
	t.Helper()
	if block.Kind != kind || block.BodyText != body || !block.ReleasedAt.Equal(releasedAt) {
		t.Fatalf("unexpected Steam block: %+v", block)
	}
}

func stringPointer(value string) *string {
	return &value
}
