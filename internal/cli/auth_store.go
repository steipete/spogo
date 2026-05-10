package cli

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/cookies"
)

func (cmd *AuthClearCmd) Run(ctx *app.Context) error {
	path := activeCookiePath(ctx)
	if path == "" {
		return fmt.Errorf("no cookie path configured")
	}
	if _, err := os.Stat(path); err == nil {
		if err := trashFile(path); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	profileCfg := ctx.Profile
	profileCfg.CookiePath = ""
	if err := ctx.SaveProfile(profileCfg); err != nil {
		return err
	}
	if err := ctx.ClearCache(); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{fmt.Sprintf("Moved %s to Trash", path)})
}

func readCookies(ctx *app.Context) ([]*http.Cookie, string, error) {
	path := activeCookiePath(ctx)
	if path != "" {
		fileCookies, err := cookies.Read(path)
		if err == nil {
			return fileCookies, "file", nil
		}
	}
	source := cookies.BrowserSource{
		Browser: defaultBrowserName(ctx.Profile.Browser),
		Profile: ctx.Profile.BrowserProfile,
		Domain:  "spotify.com",
	}
	browserCookies, err := source.Cookies(ctx.CommandContext())
	if err != nil {
		return nil, "", err
	}
	return browserCookies, "browser", nil
}

func activeCookiePath(ctx *app.Context) string {
	path := strings.TrimSpace(ctx.Profile.CookiePath)
	if path == "" {
		path = ctx.ResolveCookiePath()
	}
	return path
}

func defaultBrowserName(browser string) string {
	browser = strings.TrimSpace(browser)
	if browser == "" {
		return "chrome"
	}
	return browser
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
