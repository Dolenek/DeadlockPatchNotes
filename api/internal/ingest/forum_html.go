package ingest

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	xhtml "golang.org/x/net/html"
)

var spaceRegex = regexp.MustCompile(`\s+`)

func extractForumBody(root *xhtml.Node) (string, string, string) {
	var lines []string
	var steamURL string
	var steamImage string

	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node == nil {
			return
		}

		if node.Type == xhtml.TextNode {
			text := cleanLine(node.Data)
			if text != "" {
				lines = append(lines, text)
			}
			return
		}

		if node.Type != xhtml.ElementNode {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
			return
		}

		switch node.Data {
		case "br", "p", "div", "li", "h1", "h2", "h3", "h4", "h5", "h6":
			lines = append(lines, "")
		}

		if node.Data == "a" {
			href := strings.TrimSpace(attr(node, "href"))
			if steamURL == "" && isSteamNewsURL(href) {
				steamURL = href
			}
		}

		if node.Data == "div" && steamURL == "" {
			if candidate := strings.TrimSpace(attr(node, "data-url")); isSteamNewsURL(candidate) {
				steamURL = candidate
			}
		}

		if node.Data == "img" && steamImage == "" {
			src := strings.TrimSpace(attr(node, "src"))
			if decoded := decodeProxyImage(src); decoded != "" {
				steamImage = decoded
			} else if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
				steamImage = src
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}

		switch node.Data {
		case "p", "div", "li", "h1", "h2", "h3", "h4", "h5", "h6":
			lines = append(lines, "")
		}
	}

	walk(root)
	cleaned := compactLines(lines)

	if steamURL != "" && looksLikeSteamUnfurl(cleaned) {
		cleaned = ""
	}

	return cleaned, steamURL, steamImage
}

func decodeProxyImage(src string) string {
	if src == "" {
		return ""
	}
	parsed, err := url.Parse(src)
	if err != nil {
		return ""
	}
	q := parsed.Query().Get("image")
	if q == "" {
		return ""
	}
	decoded, err := url.QueryUnescape(q)
	if err != nil {
		return ""
	}
	return decoded
}

func isSteamNewsURL(value string) bool {
	return strings.HasPrefix(value, "https://store.steampowered.com/news/app/1422450/view/")
}

func looksLikeSteamUnfurl(body string) bool {
	if body == "" {
		return true
	}
	lower := strings.ToLower(body)
	return strings.Contains(lower, "store.steampowered.com") && len(strings.Split(body, "\n")) <= 6
}

func compactLines(lines []string) string {
	result := make([]string, 0, len(lines))
	previousBlank := false

	for _, raw := range lines {
		line := cleanLine(raw)
		if line == "" {
			if previousBlank {
				continue
			}
			if len(result) == 0 {
				continue
			}
			previousBlank = true
			result = append(result, "")
			continue
		}
		if strings.EqualFold(line, "Read more") || strings.EqualFold(line, "store.steampowered.com") {
			continue
		}
		previousBlank = false
		result = append(result, line)
	}

	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	return strings.Join(result, "\n")
}

func cleanLine(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = spaceRegex.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func findNodes(root *xhtml.Node, predicate func(*xhtml.Node) bool) []*xhtml.Node {
	results := make([]*xhtml.Node, 0, 32)
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node == nil {
			return
		}
		if predicate(node) {
			results = append(results, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return results
}

func firstNode(root *xhtml.Node, predicate func(*xhtml.Node) bool) *xhtml.Node {
	var found *xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node == nil || found != nil {
			return
		}
		if predicate(node) {
			found = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
			if found != nil {
				return
			}
		}
	}
	walk(root)
	return found
}

func attr(node *xhtml.Node, key string) string {
	for _, a := range node.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(node *xhtml.Node, classToken string) bool {
	classList := strings.Fields(attr(node, "class"))
	for _, token := range classList {
		if token == classToken {
			return true
		}
	}
	return false
}
