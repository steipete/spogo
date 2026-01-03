package cli

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
)

type AuthCmd struct {
	Status AuthStatusCmd `kong:"cmd,help='Show cookie status.'"`
	Import AuthImportCmd `kong:"cmd,help='Import browser cookies.'"`
	Clear  AuthClearCmd  `kong:"cmd,help='Clear stored cookies.'"`
}

type AuthStatusCmd struct{}

type AuthImportCmd struct {
	Browser    string `help:"Browser name (chrome|brave|edge|firefox|safari)."`
	Profile    string `name:"browser-profile" help:"Browser profile name."`
	CookiePath string `help:"Cookie cache file path."`
	Domain     string `help:"Cookie domain suffix." default:"spotify.com"`
}

type AuthClearCmd struct{}

type authStatusPayload struct {
	CookieCount int    `json:"cookie_count"`
	HasSPDC     bool   `json:"has_sp_dc"`
	Source      string `json:"source"`
}

func (cmd *AuthStatusCmd) Run(ctx *app.Context) error {
	cookiesList, sourceLabel, err := readCookies(ctx)
	if err != nil {
		return err
	}
	hasSPDC := false
	for _, cookie := range cookiesList {
		if cookie.Name == "sp_dc" {
			hasSPDC = true
			break
		}
	}
	payload := authStatusPayload{
		CookieCount: len(cookiesList),
		HasSPDC:     hasSPDC,
		Source:      sourceLabel,
	}
	plain := []string{fmt.Sprintf("%d\t%t\t%s", payload.CookieCount, payload.HasSPDC, payload.Source)}
	human := []string{fmt.Sprintf("Cookies: %d (%s)", payload.CookieCount, payload.Source)}
	if hasSPDC {
		human = append(human, "Session cookie: sp_dc")
	} else {
		human = append(human, "Session cookie: missing sp_dc")
	}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *AuthImportCmd) Run(ctx *app.Context) error {
	browser := strings.ToLower(strings.TrimSpace(cmd.Browser))
	profile := strings.TrimSpace(cmd.Profile)
	if browser == "" {
		browser = ctx.Profile.Browser
	}
	if browser == "" {
		browser = "chrome"
	}
	if profile == "" {
		profile = ctx.Profile.BrowserProfile
	}
	domain := strings.TrimSpace(cmd.Domain)
	source := cookies.BrowserSource{
		Browser: browser,
		Profile: profile,
		Domain:  domain,
	}
	if ctx.Output.Format == output.FormatHuman {
		_ = ctx.Output.WriteLines([]string{"Reading browser cookies..."})
	}
	cookiesList, err := source.Cookies(context.Background())
	if err != nil {
		return err
	}
	path := cmd.CookiePath
	if path == "" {
		path = ctx.ResolveCookiePath()
	}
	if err := cookies.Write(path, cookiesList); err != nil {
		return err
	}
	profileCfg := ctx.Profile
	profileCfg.CookiePath = path
	if browser != "" {
		profileCfg.Browser = browser
	}
	if profile != "" {
		profileCfg.BrowserProfile = profile
	}
	if err := ctx.SaveProfile(profileCfg); err != nil {
		return err
	}
	human := []string{fmt.Sprintf("Saved %d cookies to %s", len(cookiesList), path)}
	plain := []string{fmt.Sprintf("%d\t%s", len(cookiesList), path)}
	payload := map[string]any{
		"cookie_count": len(cookiesList),
		"path":         path,
	}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *AuthClearCmd) Run(ctx *app.Context) error {
	path := ctx.ResolveCookiePath()
	if path == "" {
		return fmt.Errorf("no cookie path configured")
	}
	if err := trashFile(path); err != nil {
		return err
	}
	plain := []string{"ok"}
	human := []string{fmt.Sprintf("Moved %s to Trash", path)}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, plain, human)
}

func readCookies(ctx *app.Context) ([]*http.Cookie, string, error) {
	cookiePath := ctx.Profile.CookiePath
	if cookiePath == "" {
		cookiePath = ctx.ResolveCookiePath()
	}
	if cookiePath != "" {
		fileCookies, err := cookies.Read(cookiePath)
		if err == nil {
			return fileCookies, "file", nil
		}
	}
	browser := strings.ToLower(strings.TrimSpace(ctx.Profile.Browser))
	if browser == "" {
		browser = "chrome"
	}
	browserSource := cookies.BrowserSource{Browser: browser, Profile: ctx.Profile.BrowserProfile, Domain: "spotify.com"}
	browserCookies, err := browserSource.Cookies(context.Background())
	if err != nil {
		return nil, "", err
	}
	return browserCookies, "browser", nil
}

func trashFile(path string) error {
	if _, err := exec.LookPath("trash"); err != nil {
		return fmt.Errorf("trash command not found; delete %s manually", path)
	}
	cmd := exec.Command("trash", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trash failed: %w", err)
	}
	return nil
}
