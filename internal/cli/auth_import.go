package cli

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
)

func (cmd *AuthImportCmd) Run(ctx *app.Context) error {
	source := cookies.BrowserSource{
		Browser: normalizeBrowserName(cmd.Browser, ctx.Profile.Browser),
		Profile: normalizeBrowserProfile(cmd.Profile, ctx.Profile.BrowserProfile),
		Domain:  strings.TrimSpace(cmd.Domain),
	}
	if ctx.Output.Format == output.FormatHuman {
		_ = ctx.Output.WriteLines([]string{"Reading browser cookies..."})
	}
	cookiesList, err := source.Cookies(ctx.CommandContext())
	if err != nil {
		return err
	}
	profileCfg := ctx.Profile
	profileCfg.Browser = source.Browser
	if source.Profile != "" {
		profileCfg.BrowserProfile = source.Profile
	}
	return saveCookies(ctx, cmd.CookiePath, cookiesList, profileCfg)
}

func normalizeBrowserName(primary, fallback string) string {
	browser := strings.ToLower(strings.TrimSpace(primary))
	if browser == "" {
		browser = strings.TrimSpace(fallback)
	}
	if browser == "" {
		return "chrome"
	}
	return browser
}

func normalizeBrowserProfile(primary, fallback string) string {
	profile := strings.TrimSpace(primary)
	if profile == "" {
		return strings.TrimSpace(fallback)
	}
	return profile
}

func saveCookies(ctx *app.Context, path string, cookiesList []*http.Cookie, profileCfg config.Profile) error {
	if path == "" {
		path = ctx.ResolveCookiePath()
	}
	if err := cookies.Write(path, cookiesList); err != nil {
		return err
	}
	profileCfg.CookiePath = path
	if err := ctx.SaveProfile(profileCfg); err != nil {
		return err
	}
	if err := ctx.ClearCache(); err != nil {
		return err
	}
	human := []string{fmt.Sprintf("Saved %d cookies to %s", len(cookiesList), path)}
	plain := []string{fmt.Sprintf("%d\t%s", len(cookiesList), path)}
	payload := map[string]any{"cookie_count": len(cookiesList), "path": path}
	return ctx.Output.Emit(payload, plain, human)
}
