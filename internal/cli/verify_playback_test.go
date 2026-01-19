package cli

import (
	"context"
	"testing"
	"time"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestPlayCmdVerifyPlaybackWarnOnly(t *testing.T) {
	ctx, _, errOut := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Settings.VerifyPlayback = 5 * time.Millisecond
	ctx.Settings.VerifyPlaybackFail = false

	playCalled := false
	playbackCalls := 0
	mock := &testutil.SpotifyMock{
		PlayFn: func(ctx context.Context, uri string) error {
			playCalled = true
			return nil
		},
		PlaybackFn: func(ctx context.Context) (spotify.PlaybackStatus, error) {
			playbackCalls++
			return spotify.PlaybackStatus{ProgressMS: 0}, nil
		},
	}
	ctx.SetSpotify(mock)

	cmd := PlayCmd{Item: "spotify:track:t1"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !playCalled {
		t.Fatalf("expected play")
	}
	if playbackCalls == 0 {
		t.Fatalf("expected playback polls")
	}
	if errOut.String() == "" {
		t.Fatalf("expected warning output")
	}
}

func TestPlayCmdVerifyPlaybackStrictExit(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Settings.VerifyPlayback = 5 * time.Millisecond
	ctx.Settings.VerifyPlaybackFail = true

	mock := &testutil.SpotifyMock{
		PlayFn: func(ctx context.Context, uri string) error { return nil },
		PlaybackFn: func(ctx context.Context) (spotify.PlaybackStatus, error) {
			return spotify.PlaybackStatus{ProgressMS: 0}, nil
		},
	}
	ctx.SetSpotify(mock)

	cmd := PlayCmd{Item: "spotify:track:t1"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatalf("expected error")
	}
	if app.ExitCode(err) != 5 {
		t.Fatalf("expected exit code 5, got %d", app.ExitCode(err))
	}
}

func TestDeviceSetCmdVerifyPlaybackWarnOnly(t *testing.T) {
	ctx, _, errOut := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Settings.VerifyPlayback = 5 * time.Millisecond
	ctx.Settings.VerifyPlaybackFail = false

	playbackCalls := 0
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return []spotify.Device{{ID: "d1", Name: "Desk"}}, nil
		},
		TransferFn: func(ctx context.Context, deviceID string) error {
			if deviceID != "d1" {
				t.Fatalf("device id %s", deviceID)
			}
			return nil
		},
		PlaybackFn: func(ctx context.Context) (spotify.PlaybackStatus, error) {
			playbackCalls++
			return spotify.PlaybackStatus{ProgressMS: 0}, nil
		},
	}
	ctx.SetSpotify(mock)

	cmd := DeviceSetCmd{Device: "Desk"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if playbackCalls == 0 {
		t.Fatalf("expected playback polls")
	}
	if errOut.String() == "" {
		t.Fatalf("expected warning output")
	}
}
