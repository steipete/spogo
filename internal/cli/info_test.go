package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestInfoTrackCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetTrackFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{ID: id, Name: "Song", Type: "track"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoTrackCmd{InfoArgs: InfoArgs{ID: "spotify:track:t1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestInfoTrackCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetTrackFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoTrackCmd{InfoArgs: InfoArgs{ID: "spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInfoTrackCmdInvalidID(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		GetTrackFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, nil
		},
	})
	cmd := InfoTrackCmd{InfoArgs: InfoArgs{ID: "spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}
