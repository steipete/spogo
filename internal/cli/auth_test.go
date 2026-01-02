package cli

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/browserutils/kooky"
	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/testutil"
)

func TestAuthStatusCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	path := filepath.Join(t.TempDir(), "cookies.json")
	if err := cookies.Write(path, []*http.Cookie{{Name: "sp_dc", Value: "token"}}); err != nil {
		t.Fatalf("write: %v", err)
	}
	ctx.Profile.CookiePath = path
	cmd := AuthStatusCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestAuthStatusCmdMissingSPDC(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	path := filepath.Join(t.TempDir(), "cookies.json")
	if err := cookies.Write(path, []*http.Cookie{{Name: "other", Value: "token"}}); err != nil {
		t.Fatalf("write: %v", err)
	}
	ctx.Profile.CookiePath = path
	cmd := AuthStatusCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestAuthStatusBrowserFallback(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	restore := cookies.SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{&kooky.Cookie{Cookie: http.Cookie{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}}, nil
	})
	defer restore()
	ctx.Profile.CookiePath = filepath.Join(t.TempDir(), "missing.json")
	cmd := AuthStatusCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestAuthImportCmd(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Config = config.Default()
	ctx.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	ctx.ProfileKey = "default"
	restore := cookies.SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{&kooky.Cookie{Cookie: http.Cookie{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}}, nil
	})
	defer restore()
	path := filepath.Join(t.TempDir(), "cookies.json")
	cmd := AuthImportCmd{Browser: "chrome", CookiePath: path}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := cookies.Read(path); err != nil {
		t.Fatalf("expected cookies file")
	}
}

func TestAuthImportCmdDefaultPath(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatHuman)
	ctx.Config = config.Default()
	ctx.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	ctx.ProfileKey = "default"
	ctx.Profile = config.Profile{Browser: "firefox", BrowserProfile: "Default"}
	restore := cookies.SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{&kooky.Cookie{Cookie: http.Cookie{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}}, nil
	})
	defer restore()
	cmd := AuthImportCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	path := ctx.ResolveCookiePath()
	if _, err := cookies.Read(path); err != nil {
		t.Fatalf("expected cookies file")
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

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
	ctx.ConfigPath = filepath.Join(dir, "config.toml")
	ctx.ProfileKey = "default"
	path := filepath.Join(dir, "cookies", "default.json")
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
}

func TestTrashFileMissing(t *testing.T) {
	t.Setenv("PATH", "")
	if err := trashFile("/tmp/missing"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestTrashFileSuccess(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "trash")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("PATH", dir)
	if err := trashFile("/tmp/missing"); err != nil {
		t.Fatalf("expected success")
	}
}
