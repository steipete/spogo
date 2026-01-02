package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestClampLimit(t *testing.T) {
	if clampLimit(0) != 20 {
		t.Fatalf("expected default")
	}
	if clampLimit(100) != 50 {
		t.Fatalf("expected max")
	}
	if clampLimit(10) != 10 {
		t.Fatalf("expected unchanged")
	}
}

func TestRunSearchLimitCapped(t *testing.T) {
	ctx, _, errOut := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		SearchFn: func(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
			if limit != 50 {
				t.Fatalf("limit %d", limit)
			}
			return spotify.SearchResult{Type: kind, Total: 0}, nil
		},
	}
	ctx.SetSpotify(mock)
	if err := runSearch(ctx, "track", SearchArgs{Query: "song", Limit: 100}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(errOut.String(), "limit capped") {
		t.Fatalf("expected cap warning")
	}
}

func TestSearchAlbumCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		SearchFn: func(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
			return spotify.SearchResult{Type: kind, Total: 0}, nil
		},
	})
	cmd := SearchAlbumCmd{SearchArgs: SearchArgs{Query: "album"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestSearchArtistCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		SearchFn: func(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
			return spotify.SearchResult{Type: kind, Total: 0}, nil
		},
	})
	cmd := SearchArtistCmd{SearchArgs: SearchArgs{Query: "artist"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestSearchPlaylistCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		SearchFn: func(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
			return spotify.SearchResult{Type: kind, Total: 0}, nil
		},
	})
	cmd := SearchPlaylistCmd{SearchArgs: SearchArgs{Query: "playlist"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestSearchEpisodeCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		SearchFn: func(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
			return spotify.SearchResult{Type: kind, Total: 0}, nil
		},
	})
	cmd := SearchEpisodeCmd{SearchArgs: SearchArgs{Query: "episode"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestSearchShowCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		SearchFn: func(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
			return spotify.SearchResult{Type: kind, Total: 0}, nil
		},
	})
	cmd := SearchShowCmd{SearchArgs: SearchArgs{Query: "show"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}
