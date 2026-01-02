package cli

import (
	"context"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestInfoAlbumCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetAlbumFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{ID: id, Name: "Album", Type: "album"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoAlbumCmd{InfoArgs: InfoArgs{ID: "spotify:album:a1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestInfoArtistCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetArtistFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{ID: id, Name: "Artist", Type: "artist"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoArtistCmd{InfoArgs: InfoArgs{ID: "spotify:artist:a1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestInfoPlaylistCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetPlaylistFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{ID: id, Name: "Playlist", Type: "playlist"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoPlaylistCmd{InfoArgs: InfoArgs{ID: "spotify:playlist:p1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestInfoShowCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetShowFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{ID: id, Name: "Show", Type: "show"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoShowCmd{InfoArgs: InfoArgs{ID: "spotify:show:s1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestInfoEpisodeCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetEpisodeFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{ID: id, Name: "Episode", Type: "episode"}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoEpisodeCmd{InfoArgs: InfoArgs{ID: "spotify:episode:e1"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestInfoAlbumCmdInvalidID(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetAlbumFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := InfoAlbumCmd{InfoArgs: InfoArgs{ID: "spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInfoArtistCmdInvalidID(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		GetArtistFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, nil
		},
	})
	cmd := InfoArtistCmd{InfoArgs: InfoArgs{ID: "spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInfoPlaylistCmdInvalidID(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		GetPlaylistFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, nil
		},
	})
	cmd := InfoPlaylistCmd{InfoArgs: InfoArgs{ID: "spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInfoShowCmdInvalidID(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		GetShowFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, nil
		},
	})
	cmd := InfoShowCmd{InfoArgs: InfoArgs{ID: "spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInfoEpisodeCmdInvalidID(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		GetEpisodeFn: func(ctx context.Context, id string) (spotify.Item, error) {
			return spotify.Item{}, nil
		},
	})
	cmd := InfoEpisodeCmd{InfoArgs: InfoArgs{ID: "spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}
