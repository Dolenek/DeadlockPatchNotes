package httpapi

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

const apiContentSecurityPolicy = "default-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'"

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", apiContentSecurityPolicy)
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(), payment=(), usb=()")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func httpsRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstForwardedProtocol(r.Header.Get("X-Forwarded-Proto")) != "http" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Vary", "X-Forwarded-Proto")
		http.Redirect(w, r, httpsRedirectURL(r), http.StatusPermanentRedirect)
	})
}

func firstForwardedProtocol(raw string) string {
	first, _, _ := strings.Cut(raw, ",")
	return strings.ToLower(strings.TrimSpace(first))
}

func httpsRedirectURL(r *http.Request) string {
	base := normalizeSiteURL(os.Getenv("SITE_URL"))
	if !strings.HasPrefix(base, "https://") {
		base = canonicalSiteURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		parsed, _ = url.Parse(canonicalSiteURL)
	}
	parsed.Path = r.URL.Path
	parsed.RawPath = r.URL.RawPath
	parsed.RawQuery = r.URL.RawQuery
	parsed.Fragment = ""
	return parsed.String()
}
