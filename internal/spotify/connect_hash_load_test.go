package spotify

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestHashResolverLoad(t *testing.T) {
	hash := strings.Repeat("a", 64)
	mainJS := `var a={1:"web-player/main"};var b={1:"abcdef"};`
	chunkBody := `searchDesktop blah sha256Hash":"` + hash + `"`
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
	got, err := resolver.Hash(context.Background(), "searchDesktop")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if got != hash {
		t.Fatalf("unexpected hash: %s", got)
	}
}
