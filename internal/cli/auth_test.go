package cli

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/testutil"
	"github.com/steipete/sweetcookie"
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
	restore := cookies.SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}}, nil
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
	restore := cookies.SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}}, nil
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
	restore := cookies.SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}}, nil
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

func TestGlobalsSettingsPassesNoInput(t *testing.T) {
	settings, err := (Globals{NoInput: true}).Settings()
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if !settings.NoInput {
		t.Fatalf("expected no_input true")
	}
}

func TestGlobalsSettingsRejectsPlainAndJSON(t *testing.T) {
	_, err := (Globals{JSON: true, Plain: true}).Settings()
	if err == nil {
		t.Fatalf("expected error")
	}
}
