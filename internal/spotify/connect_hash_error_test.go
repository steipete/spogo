package spotify

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestPickWebPlayerBundleMissing(t *testing.T) {
	if _, err := pickWebPlayerBundle("<html></html>"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseWebpackMapsNoMaps(t *testing.T) {
	if _, _, err := parseWebpackMaps("var a = 1;"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestHashResolverLoadMissingHash(t *testing.T) {
	mainJS := `var a={1:"web-player/main"};var b={1:"abcdef"};`
	chunkBody := `something without hash`
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Host == "open.spotify.com":
			html := `<script src="https://open.spotifycdn.com/cdn/build/web-player/main.js"></script>`
			return textResponse(http.StatusOK, html), nil
		case strings.Contains(req.URL.Path, "/web-player/main.js"):
			return textResponse(http.StatusOK, mainJS), nil
		case strings.Contains(req.URL.Path, "web-player/main.abcdef.js"):
			return textResponse(http.StatusOK, chunkBody), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	client := &http.Client{Transport: transport}
	resolver := newHashResolver(client, &connectSession{client: client})
	if _, err := resolver.Hash(context.Background(), "searchDesktop"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestHashResolverHTTPError(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return textResponse(http.StatusInternalServerError, "fail"), nil
	})
	client := &http.Client{Transport: transport}
	resolver := newHashResolver(client, &connectSession{client: client})
	if _, err := resolver.fetchWebPlayerHTML(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := resolver.fetchText(context.Background(), "https://open.spotifycdn.com/cdn/build/web-player/main.js"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestFilterMissing(t *testing.T) {
	remaining := filterMissing([]string{"a", "b"}, map[string]string{"a": "hash"})
	if len(remaining) != 1 || remaining[0] != "b" {
		t.Fatalf("unexpected remaining: %#v", remaining)
	}
}

func TestBundleBaseURL(t *testing.T) {
	base := bundleBaseURL("https://open.spotifycdn.com/cdn/build/mobile-web-player/main.js")
	if !strings.HasSuffix(base, "/mobile-web-player/") {
		t.Fatalf("unexpected base: %s", base)
	}
	if bundleBaseURL("main.js") == "" {
		t.Fatalf("expected fallback base")
	}
}
