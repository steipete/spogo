package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestLibraryTracksAddCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		LibraryModifyFn: func(ctx context.Context, path string, ids []string, method string) error {
			return errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := LibraryTracksAddCmd{IDs: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryTracksRemoveCmdInvalid(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := LibraryTracksRemoveCmd{IDs: []string{"spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryTracksRemoveCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		LibraryModifyFn: func(ctx context.Context, path string, ids []string, method string) error {
			return errors.New("boom")
		},
	})
	cmd := LibraryTracksRemoveCmd{IDs: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryAlbumsAddCmdInvalid(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := LibraryAlbumsAddCmd{IDs: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryAlbumsAddCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		LibraryModifyFn: func(ctx context.Context, path string, ids []string, method string) error {
			return errors.New("boom")
		},
	})
	cmd := LibraryAlbumsAddCmd{IDs: []string{"spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryAlbumsRemoveCmdInvalid(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := LibraryAlbumsRemoveCmd{IDs: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryAlbumsRemoveCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		LibraryModifyFn: func(ctx context.Context, path string, ids []string, method string) error {
			return errors.New("boom")
		},
	})
	cmd := LibraryAlbumsRemoveCmd{IDs: []string{"spotify:album:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryArtistsFollowCmdInvalid(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := LibraryArtistsFollowCmd{IDs: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryArtistsFollowCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		FollowArtistsFn: func(ctx context.Context, ids []string, method string) error {
			return errors.New("boom")
		},
	})
	cmd := LibraryArtistsFollowCmd{IDs: []string{"spotify:artist:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryArtistsUnfollowCmdInvalid(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := LibraryArtistsUnfollowCmd{IDs: []string{"spotify:track:t1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryArtistsUnfollowCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.SetSpotify(&testutil.SpotifyMock{
		FollowArtistsFn: func(ctx context.Context, ids []string, method string) error {
			return errors.New("boom")
		},
	})
	cmd := LibraryArtistsUnfollowCmd{IDs: []string{"spotify:artist:a1"}}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLibraryArtistsListCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		FollowedArtistsFn: func(ctx context.Context, limit int, after string) ([]spotify.Item, int, string, error) {
			return nil, 0, "", errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := LibraryArtistsListCmd{Limit: 1}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}
