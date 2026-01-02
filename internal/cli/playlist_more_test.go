package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestPlaylistAddCmdInvalidPlaylist(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{AddTracksFn: func(ctx context.Context, playlistID string, uris []string) error { return nil }})
	cmd := PlaylistAddCmd{Playlist: "spotify:track:t1", Tracks: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPlaylistAddCmdInvalidTrack(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{AddTracksFn: func(ctx context.Context, playlistID string, uris []string) error { return nil }})
	cmd := PlaylistAddCmd{Playlist: "spotify:playlist:p1", Tracks: []string{"spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPlaylistRemoveCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		RemoveTracksFn: func(ctx context.Context, playlistID string, uris []string) error {
			return errors.New("boom")
		},
	})
	cmd := PlaylistRemoveCmd{Playlist: "spotify:playlist:p1", Tracks: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPlaylistTracksCmdHumanHeader(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatHuman)
	ctx.SetSpotify(&testutil.SpotifyMock{
		PlaylistTracksFn: func(ctx context.Context, id string, limit, offset int) ([]spotify.Item, int, error) {
			return []spotify.Item{{ID: "t1", Name: "Track", Type: "track"}}, 1, nil
		},
	})
	cmd := PlaylistTracksCmd{Playlist: "spotify:playlist:p1", Limit: 1}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), "Tracks:") {
		t.Fatalf("expected header")
	}
}

func TestPlaylistCreateCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		CreatePlaylistFn: func(ctx context.Context, name string, public, collaborative bool) (spotify.Item, error) {
			return spotify.Item{}, errors.New("boom")
		},
	})
	cmd := PlaylistCreateCmd{Name: "Fail"}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPlaylistTracksCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		PlaylistTracksFn: func(ctx context.Context, id string, limit, offset int) ([]spotify.Item, int, error) {
			return nil, 0, errors.New("boom")
		},
	})
	cmd := PlaylistTracksCmd{Playlist: "spotify:playlist:p1", Limit: 1}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}
