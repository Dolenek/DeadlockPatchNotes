package ingest

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
)

var (
	threadHrefRegex = regexp.MustCompile(`href="(/threads/[^"]+/)"`)
	nextPageRegex   = regexp.MustCompile(`rel="next"\s+href="([^"]+)"`)
)

type ForumThreadRef struct {
	URL string
}

type ForumPost struct {
	PostID       string
	PublishedAt  time.Time
	BodyText     string
	SteamURL     string
	SteamImage   string
	ForumPostURL string
}

type ForumThread struct {
	ThreadID int64
	Slug     string
	Title    string
	URL      string
	Posts    []ForumPost
}

func CrawlChangelogThreads(ctx context.Context, client *http.Client, changelogURL string, maxPages int) ([]ForumThreadRef, error) {
	if maxPages <= 0 {
		maxPages = 20
	}

	seenPages := map[string]bool{}
	seenThreads := map[string]bool{}
	refs := make([]ForumThreadRef, 0, 128)

	current := changelogURL
	for i := 0; i < maxPages && current != ""; i++ {
		if seenPages[current] {
			break
		}
		seenPages[current] = true

		raw, err := fetchText(ctx, client, current)
		if err != nil {
			return nil, err
		}

		for _, match := range threadHrefRegex.FindAllStringSubmatch(raw, -1) {
			if len(match) < 2 {
				continue
			}
			abs := resolveURL(current, match[1])
			if !isPatchThreadURL(abs) {
				continue
			}
			if seenThreads[abs] {
				continue
			}
			seenThreads[abs] = true
			refs = append(refs, ForumThreadRef{URL: abs})
		}

		nextMatch := nextPageRegex.FindStringSubmatch(raw)
		if len(nextMatch) < 2 {
			break
		}
		next := resolveURL(current, nextMatch[1])
		if next == current {
			break
		}
		current = next
	}

	sort.SliceStable(refs, func(i, j int) bool {
		return refs[i].URL < refs[j].URL
	})
	return refs, nil
}

func FetchThread(ctx context.Context, client *http.Client, threadURL string) (ForumThread, error) {
	raw, err := fetchText(ctx, client, threadURL)
	if err != nil {
		return ForumThread{}, err
	}

	threadID, slug, err := parseThreadIdentity(threadURL)
	if err != nil {
		return ForumThread{}, err
	}

	title := parseThreadTitle(raw)
	if title == "" {
		title = strings.ReplaceAll(slug, "-", " ")
	}

	doc, err := xhtml.Parse(strings.NewReader(raw))
	if err != nil {
		return ForumThread{}, fmt.Errorf("parse html %s: %w", threadURL, err)
	}

	articles := findNodes(doc, func(n *xhtml.Node) bool {
		if n.Type != xhtml.ElementNode || n.Data != "article" {
			return false
		}
		return hasClass(n, "message--post") && attr(n, "data-author") == "Yoshi"
	})

	posts := make([]ForumPost, 0, len(articles))
	for _, article := range articles {
		postID := strings.TrimPrefix(attr(article, "data-content"), "post-")
		if postID == "" {
			continue
		}

		timeNode := firstNode(article, func(n *xhtml.Node) bool {
			return n.Type == xhtml.ElementNode && n.Data == "time" && attr(n, "datetime") != ""
		})
		if timeNode == nil {
			continue
		}
		publishedAt, err := parseForumTime(attr(timeNode, "datetime"))
		if err != nil {
			continue
		}

		bbWrapper := firstNode(article, func(n *xhtml.Node) bool {
			return n.Type == xhtml.ElementNode && n.Data == "div" && hasClass(n, "bbWrapper")
		})
		if bbWrapper == nil {
			continue
		}

		bodyText, steamURL, steamImage := extractForumBody(bbWrapper)
		postURL := threadURL
		if !strings.HasSuffix(postURL, "/") {
			postURL += "/"
		}
		postURL += "post-" + postID

		if bodyText == "" && steamURL == "" {
			continue
		}

		posts = append(posts, ForumPost{
			PostID:       postID,
			PublishedAt:  publishedAt,
			BodyText:     bodyText,
			SteamURL:     steamURL,
			SteamImage:   steamImage,
			ForumPostURL: postURL,
		})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PublishedAt.Before(posts[j].PublishedAt)
	})

	return ForumThread{
		ThreadID: threadID,
		Slug:     slug,
		Title:    title,
		URL:      threadURL,
		Posts:    posts,
	}, nil
}

func fetchText(ctx context.Context, client *http.Client, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request %s: %w", targetURL, err)
	}
	req.Header.Set("User-Agent", "deadlockpatchnotes-ingester/1.0")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request %s: %w", targetURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("request %s: status %d", targetURL, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response %s: %w", targetURL, err)
	}
	return string(body), nil
}

func parseForumTime(value string) (time.Time, error) {
	// Example: 2026-03-06T13:37:48-0800
	return time.Parse("2006-01-02T15:04:05-0700", value)
}

func parseThreadTitle(raw string) string {
	open := strings.Index(raw, "<title>")
	close := strings.Index(raw, "</title>")
	if open == -1 || close == -1 || close <= open+7 {
		return ""
	}
	title := strings.TrimSpace(html.UnescapeString(raw[open+7 : close]))
	title = strings.TrimSuffix(title, " | Deadlock")
	return strings.TrimSpace(title)
}

func parseThreadIdentity(threadURL string) (int64, string, error) {
	parsed, err := url.Parse(threadURL)
	if err != nil {
		return 0, "", err
	}
	base := path.Base(strings.TrimSuffix(parsed.Path, "/"))
	dot := strings.LastIndex(base, ".")
	if dot <= 0 || dot >= len(base)-1 {
		return 0, "", fmt.Errorf("invalid thread path: %s", parsed.Path)
	}
	slug := base[:dot]
	threadID, err := strconv.ParseInt(base[dot+1:], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid thread id in %s: %w", parsed.Path, err)
	}
	return threadID, slug, nil
}

func resolveURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	rel, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(rel).String()
}

func isPatchThreadURL(targetURL string) bool {
	if strings.Contains(targetURL, "changelog-feedback-process") {
		return false
	}
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	if !strings.HasPrefix(parsed.Path, "/threads/") {
		return false
	}
	return strings.Contains(parsed.Path, "-update")
}
