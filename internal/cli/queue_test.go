package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestQueueShowCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		QueueFn: func(ctx context.Context) (spotify.Queue, error) {
			return spotify.Queue{Queue: []spotify.Item{{ID: "t1", Name: "Song", Type: "track"}}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := QueueShowCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestQueueShowWithCurrent(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatHuman)
	mock := &testutil.SpotifyMock{
		QueueFn: func(ctx context.Context) (spotify.Queue, error) {
			item := spotify.Item{ID: "t1", Name: "Song", Type: "track"}
			return spotify.Queue{CurrentlyPlaying: &item, Queue: []spotify.Item{}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := QueueShowCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestQueueAddCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		QueueAddFn: func(ctx context.Context, uri string) error {
			if uri != "spotify:track:t1" {
				t.Fatalf("uri %s", uri)
			}
			return nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := QueueAddCmd{Item: "spotify:track:t1"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestQueueAddCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		QueueAddFn: func(ctx context.Context, uri string) error {
			return errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := QueueAddCmd{Item: "spotify:track:t1"}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestQueueAddInvalid(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := QueueAddCmd{Item: "spotify:album:a1"}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestQueueClearCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := QueueClearCmd{}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}
