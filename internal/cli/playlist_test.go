package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestPlaylistAddCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	called := false
	mock := &testutil.SpotifyMock{
		AddTracksFn: func(ctx context.Context, playlistID string, uris []string) error {
			called = true
			if playlistID != "p1" {
				t.Fatalf("playlist id %s", playlistID)
			}
			if len(uris) != 1 || uris[0] != "spotify:track:t1" {
				t.Fatalf("uris: %#v", uris)
			}
			return nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := PlaylistAddCmd{Playlist: "spotify:playlist:p1", Tracks: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !called {
		t.Fatalf("expected call")
	}
}

func TestPlaylistAddCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		AddTracksFn: func(ctx context.Context, playlistID string, uris []string) error {
			return errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := PlaylistAddCmd{Playlist: "spotify:playlist:p1", Tracks: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPlaylistCreateCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		CreatePlaylistFn: func(ctx context.Context, name string, public, collaborative bool) (spotify.Item, error) {
			return spotify.Item{ID: "p1", Name: name, Type: "playlist"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := PlaylistCreateCmd{Name: "Road Trip"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestPlaylistTracksCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		PlaylistTracksFn: func(ctx context.Context, id string, limit, offset int) ([]spotify.Item, int, error) {
			return []spotify.Item{{ID: "t1", Name: "Track", Type: "track"}}, 1, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := PlaylistTracksCmd{Playlist: "spotify:playlist:p1", Limit: 1}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestPlaylistRemoveCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		RemoveTracksFn: func(ctx context.Context, playlistID string, uris []string) error {
			if playlistID != "p1" {
				t.Fatalf("playlist %s", playlistID)
			}
			return nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := PlaylistRemoveCmd{Playlist: "spotify:playlist:p1", Tracks: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}
