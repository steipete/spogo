package cli

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/testutil"
	"github.com/steipete/sweetcookie"
)

func withStdin(t *testing.T, contents string, fn func()) {
	t.Helper()
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString(contents); err != nil {
		_ = r.Close()
		_ = w.Close()
		t.Fatalf("write: %v", err)
	}
	_ = w.Close()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
	fn()
	_ = r.Close()
}

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

func TestNormalizeCookieDomain(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ".spotify.com"},
		{"spotify.com", ".spotify.com"},
		{".spotify.com", ".spotify.com"},
		{"https://open.spotify.com/", ".open.spotify.com"},
	}
	for _, tc := range cases {
		if got := normalizeCookieDomain(tc.in); got != tc.want {
			t.Fatalf("normalizeCookieDomain(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeCookieValue(t *testing.T) {
	if got, ok := extractNamedCookieValue("sp_dc=token; Path=/; Secure", "sp_dc"); !ok || got != "token" {
		t.Fatalf("expected token, got %q", got)
	}
	if got := normalizePromptCookieValue("\"token\"", "sp_dc"); got != "token" {
		t.Fatalf("expected token, got %q", got)
	}
	if got := normalizePromptCookieValue("token", "sp_dc"); got != "token" {
		t.Fatalf("expected token, got %q", got)
	}
}

func TestReadPromptCookieValueEOF(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("sp_dc=token"))
	value, err := readPromptCookieValue(reader, nil, "sp_dc", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "token" {
		t.Fatalf("expected token, got %q", value)
	}
}

func TestReadPromptCookieValueRequiredRejectsEmpty(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	_, err := readPromptCookieValue(reader, nil, "sp_dc", true)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParsePastedCookiesAnyOrder(t *testing.T) {
	values, err := parsePastedCookies(strings.NewReader("sp_t=device\nsp_dc=token\nsp_key=key\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values.spdc != "token" || values.spkey != "key" || values.spt != "device" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestAuthPasteCmdFromStdin(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Config = config.Default()
	ctx.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	ctx.ProfileKey = "default"

	dest := filepath.Join(t.TempDir(), "out.json")
	withStdin(t, "sp_t=device\nsp_dc=token\nsp_key=key\n", func() {
		cmd := AuthPasteCmd{
			CookiePath: dest,
			Domain:     "spotify.com",
		}
		if err := cmd.Run(ctx); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	cookiesList, err := cookies.Read(dest)
	if err != nil {
		t.Fatalf("read cookies: %v", err)
	}
	if len(cookiesList) != 3 {
		t.Fatalf("expected 3 cookies, got %d", len(cookiesList))
	}
	if ctx.Profile.CookiePath != dest {
		t.Fatalf("expected profile cookie path %s, got %s", dest, ctx.Profile.CookiePath)
	}
}

func TestAuthPasteCmdWarnsWhenMissingSPT(t *testing.T) {
	ctx, _, errOut := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Config = config.Default()
	ctx.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	ctx.ProfileKey = "default"
	ctx.Profile = config.Profile{Engine: "connect"}

	dest := filepath.Join(t.TempDir(), "out.json")
	withStdin(t, "sp_dc=token\n", func() {
		cmd := AuthPasteCmd{CookiePath: dest}
		if err := cmd.Run(ctx); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	if !strings.Contains(errOut.String(), "missing sp_t") {
		t.Fatalf("expected warning in stderr, got %q", errOut.String())
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

func TestAuthPasteCmdRequiresSPDC(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	withStdin(t, "", func() {
		cmd := AuthPasteCmd{}
		if err := cmd.Run(ctx); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestAuthPasteCmdNoInputFromStdin(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	ctx.Config = config.Default()
	ctx.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	ctx.ProfileKey = "default"
	ctx.Settings.NoInput = true

	dest := filepath.Join(t.TempDir(), "out.json")
	withStdin(t, "sp_dc=token\nsp_t=device\n", func() {
		cmd := AuthPasteCmd{CookiePath: dest}
		if err := cmd.Run(ctx); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	if _, err := cookies.Read(dest); err != nil {
		t.Fatalf("expected cookies file")
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
