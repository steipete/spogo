package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestDeviceSetCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	called := false
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return []spotify.Device{{ID: "d1", Name: "Desk"}}, nil
		},
		TransferFn: func(ctx context.Context, deviceID string) error {
			called = true
			if deviceID != "d1" {
				t.Fatalf("device id %s", deviceID)
			}
			return nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := DeviceSetCmd{Device: "Desk"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !called {
		t.Fatalf("expected transfer")
	}
}

func TestDeviceSetCmdSave(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")

	cfg := config.Default()
	profileKey := "p1"
	cfg.SetProfile(profileKey, config.Profile{})

	ctx.Config = cfg
	ctx.ConfigPath = configPath
	ctx.ProfileKey = profileKey
	ctx.Profile = cfg.Profile(profileKey)

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
	}
	ctx.SetSpotify(mock)
	cmd := DeviceSetCmd{Device: "Desk", Save: true}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config not written: %v", err)
	}
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	got := loaded.Profile(profileKey).Device
	if got != "d1" {
		t.Fatalf("expected saved device d1, got %q", got)
	}
}

func TestDeviceSetCmdTransferError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return []spotify.Device{{ID: "d1", Name: "Desk"}}, nil
		},
		TransferFn: func(ctx context.Context, deviceID string) error {
			return errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := DeviceSetCmd{Device: "Desk"}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDeviceListCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return []spotify.Device{{ID: "d1", Name: "Desk", Active: true}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := DeviceListCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestDeviceShowCmdNoTarget(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return []spotify.Device{
				{ID: "d1", Name: "Desk", Active: true},
				{ID: "d2", Name: "Phone"},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := DeviceShowCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestDeviceShowCmdWithTarget(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Profile.Device = "Phone"
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return []spotify.Device{
				{ID: "d1", Name: "Desk", Active: true},
				{ID: "d2", Name: "Phone"},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := DeviceShowCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestDeviceClearCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")

	cfg := config.Default()
	profileKey := "p1"
	cfg.SetProfile(profileKey, config.Profile{Device: "d1"})

	ctx.Config = cfg
	ctx.ConfigPath = configPath
	ctx.ProfileKey = profileKey
	ctx.Profile = cfg.Profile(profileKey)

	cmd := DeviceClearCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	got := loaded.Profile(profileKey).Device
	if got != "" {
		t.Fatalf("expected cleared device, got %q", got)
	}
}

func TestDeviceListCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		DevicesFn: func(ctx context.Context) ([]spotify.Device, error) {
			return nil, errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := DeviceListCmd{}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestActiveMarker(t *testing.T) {
	if activeMarker(false) != "" {
		t.Fatalf("expected empty")
	}
	if activeMarker(true) == "" {
		t.Fatalf("expected marker")
	}
}
