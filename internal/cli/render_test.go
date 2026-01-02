package cli

import (
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestRenderItemsHuman(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatHuman)
	items := []spotify.Item{{ID: "t1", Name: "Song", Type: "track", Artists: []string{"Artist"}, Album: "Album"}}
	plain, human := renderItems(ctx.Output, items)
	if len(plain) == 0 || len(human) == 0 {
		t.Fatalf("expected lines")
	}
	_ = ctx.Output.Emit(items, plain, human)
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestPlaybackFormatting(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatHuman)
	status := spotify.PlaybackStatus{IsPlaying: true, ProgressMS: 120000, Device: spotify.Device{Name: "Desk"}, Item: &spotify.Item{Name: "Song", Artists: []string{"Artist"}}}
	if playbackPlain(status) == "" {
		t.Fatalf("expected plain")
	}
	if playbackHuman(ctx.Output, status) == "" {
		t.Fatalf("expected human")
	}
}
