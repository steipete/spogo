//go:build !darwin

package app

import (
	"strings"
	"testing"

	"github.com/steipete/spogo/internal/config"
)

func TestSpotifyAppleScriptEngine_NonDarwin(t *testing.T) {
	t.Parallel()

	ctx := &Context{Profile: config.Profile{CookiePath: "/tmp/cookies.json", Engine: "applescript"}}
	if _, err := ctx.Spotify(); err == nil || !strings.Contains(err.Error(), "only available on macOS") {
		t.Fatalf("expected macOS-only error, got: %v", err)
	}
	if ctx.spotifyClient != nil {
		t.Fatalf("expected no cached client on error")
	}
}
