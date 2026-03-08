package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/testutil"
)

func TestAuthClearCmdNoPath(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := AuthClearCmd{}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAuthClearCmdSuccess(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	dir := t.TempDir()
	ctx.Config = config.Default()
	ctx.ConfigPath = filepath.Join(dir, "config.toml")
	ctx.ProfileKey = "default"
	path := filepath.Join(dir, "cookies", "default.json")
	ctx.Profile.CookiePath = path
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("[]"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	script := filepath.Join(dir, "trash")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("PATH", dir)
	cmd := AuthClearCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if ctx.Profile.CookiePath != "" {
		t.Fatalf("expected profile cookie path cleared, got %q", ctx.Profile.CookiePath)
	}
}

func TestTrashFileMissing(t *testing.T) {
	t.Setenv("PATH", "")
	if err := trashFile("/tmp/missing"); err == nil {
		t.Fatalf("expected error")
	}
}
