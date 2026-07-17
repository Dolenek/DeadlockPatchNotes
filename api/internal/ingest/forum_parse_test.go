package ingest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchThreadKeepsValidOfficialPostsInChronologicalOrder(t *testing.T) {
	markup := `<html><head><title>July Update | Deadlock</title></head><body>
		<article class="message--post" data-author="Yoshi" data-content="post-2">
			<time datetime="2026-07-02T12:00:00+0000"></time><div class="bbWrapper"><p>- Later</p></div>
		</article>
		<article class="message--post" data-author="Other" data-content="post-ignored">
			<time datetime="2026-07-01T10:00:00+0000"></time><div class="bbWrapper"><p>- Ignore</p></div>
		</article>
		<article class="message--post" data-author="Yoshi" data-content="post-invalid">
			<time datetime="invalid"></time><div class="bbWrapper"><p>- Invalid</p></div>
		</article>
		<article class="message--post" data-author="Yoshi" data-content="post-1">
			<time datetime="2026-07-01T12:00:00+0000"></time><div class="bbWrapper"><p>- Earlier</p></div>
		</article>
	</body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(markup)) }))
	defer server.Close()

	thread, err := FetchThread(t.Context(), server.Client(), server.URL+"/threads/july-update.123/")
	if err != nil {
		t.Fatalf("fetch thread: %v", err)
	}
	if thread.ThreadID != 123 || thread.Slug != "july-update" || thread.Title != "July Update" {
		t.Fatalf("unexpected thread identity: %+v", thread)
	}
	if len(thread.Posts) != 2 || thread.Posts[0].PostID != "1" || thread.Posts[1].PostID != "2" {
		t.Fatalf("unexpected official posts: %+v", thread.Posts)
	}
	if !strings.HasSuffix(thread.Posts[0].ForumPostURL, "/post-1") {
		t.Fatalf("unexpected post URL: %q", thread.Posts[0].ForumPostURL)
	}
}

func TestParseThreadIdentityRejectsMalformedPaths(t *testing.T) {
	for _, target := range []string{
		"https://forum.test/threads/no-id/",
		"https://forum.test/threads/update.not-a-number/",
		"https://forum.test/threads/.123/",
	} {
		if _, _, err := parseThreadIdentity(target); err == nil {
			t.Fatalf("expected invalid identity error for %s", target)
		}
	}
}
