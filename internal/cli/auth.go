package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
)

type AuthCmd struct {
	Status AuthStatusCmd `kong:"cmd,help='Show cookie status.'"`
	Import AuthImportCmd `kong:"cmd,help='Import browser cookies.'"`
	Paste  AuthPasteCmd  `kong:"cmd,help='Paste cookie values from the browser.'"`
	Clear  AuthClearCmd  `kong:"cmd,help='Clear stored cookies.'"`
}

type AuthStatusCmd struct{}

type AuthImportCmd struct {
	Browser    string `help:"Browser name (chrome|brave|edge|firefox|safari)."`
	Profile    string `name:"browser-profile" help:"Browser profile name."`
	CookiePath string `help:"Cookie cache file path."`
	Domain     string `help:"Cookie domain suffix." default:"spotify.com"`
}

type AuthPasteCmd struct {
	CookiePath string `help:"Cookie cache file path."`
	Domain     string `help:"Cookie domain suffix." default:"spotify.com"`
	Path       string `help:"Cookie path." default:"/"`
}

type AuthClearCmd struct{}

type authStatusPayload struct {
	CookieCount int    `json:"cookie_count"`
	HasSPDC     bool   `json:"has_sp_dc"`
	HasSPT      bool   `json:"has_sp_t"`
	HasSPKey    bool   `json:"has_sp_key"`
	Source      string `json:"source"`
}

func (cmd *AuthStatusCmd) Run(ctx *app.Context) error {
	cookiesList, sourceLabel, err := readCookies(ctx)
	if err != nil {
		return err
	}
	hasSPDC := false
	hasSPT := false
	hasSPKey := false
	for _, cookie := range cookiesList {
		switch cookie.Name {
		case "sp_dc":
			hasSPDC = true
		case "sp_t":
			hasSPT = true
		case "sp_key":
			hasSPKey = true
		}
	}
	payload := authStatusPayload{
		CookieCount: len(cookiesList),
		HasSPDC:     hasSPDC,
		HasSPT:      hasSPT,
		HasSPKey:    hasSPKey,
		Source:      sourceLabel,
	}
	plain := []string{fmt.Sprintf("%d\t%t\t%t\t%t\t%s", payload.CookieCount, payload.HasSPDC, payload.HasSPT, payload.HasSPKey, payload.Source)}
	human := []string{fmt.Sprintf("Cookies: %d (%s)", payload.CookieCount, payload.Source)}
	if hasSPDC {
		human = append(human, "Session cookie: sp_dc")
	} else {
		human = append(human, "Session cookie: missing sp_dc")
	}
	if hasSPT {
		human = append(human, "Device cookie: sp_t")
	} else {
		human = append(human, "Device cookie: missing sp_t (needed for connect playback)")
	}
	if hasSPKey {
		human = append(human, "Optional cookie: sp_key")
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
	cookiesList, err := source.Cookies(ctx.CommandContext())
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

func (cmd *AuthPasteCmd) Run(ctx *app.Context) error {
	stdinIsTTY := isatty.IsTerminal(os.Stdin.Fd())
	if ctx.Settings.NoInput && stdinIsTTY {
		return errors.New("--no-input set; pipe cookie values via stdin")
	}
	values, err := readPastedCookies(os.Stdin, ctx.Output, stdinIsTTY && !ctx.Settings.NoInput)
	if err != nil {
		return err
	}
	if values.spdc == "" {
		return errors.New("sp_dc required")
	}

	domain := normalizeCookieDomain(cmd.Domain)
	path := normalizeCookiePath(cmd.Path)

	cookiesList := []*http.Cookie{{
		Name:     "sp_dc",
		Value:    values.spdc,
		Domain:   domain,
		Path:     path,
		Secure:   true,
		HttpOnly: true,
	}}
	if values.spkey != "" {
		cookiesList = append(cookiesList, &http.Cookie{
			Name:     "sp_key",
			Value:    values.spkey,
			Domain:   domain,
			Path:     path,
			Secure:   true,
			HttpOnly: true,
		})
	}
	if values.spt != "" {
		cookiesList = append(cookiesList, &http.Cookie{
			Name:     "sp_t",
			Value:    values.spt,
			Domain:   domain,
			Path:     path,
			Secure:   true,
			HttpOnly: true,
		})
	} else if strings.EqualFold(strings.TrimSpace(ctx.Profile.Engine), "") || strings.EqualFold(strings.TrimSpace(ctx.Profile.Engine), "connect") || strings.EqualFold(strings.TrimSpace(ctx.Profile.Engine), "auto") {
		_, _ = fmt.Fprintln(ctx.Output.Err, "warning: missing sp_t; playback may fail (grab sp_t from DevTools)")
	}

	pathOut := cmd.CookiePath
	if pathOut == "" {
		pathOut = ctx.ResolveCookiePath()
	}
	if err := cookies.Write(pathOut, cookiesList); err != nil {
		return err
	}
	profileCfg := ctx.Profile
	profileCfg.CookiePath = pathOut
	if err := ctx.SaveProfile(profileCfg); err != nil {
		return err
	}
	human := []string{fmt.Sprintf("Saved %d cookies to %s", len(cookiesList), pathOut)}
	plain := []string{fmt.Sprintf("%d\t%s", len(cookiesList), pathOut)}
	payload := map[string]any{
		"cookie_count": len(cookiesList),
		"path":         pathOut,
	}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *AuthClearCmd) Run(ctx *app.Context) error {
	path := strings.TrimSpace(ctx.Profile.CookiePath)
	if path == "" {
		path = ctx.ResolveCookiePath()
	}
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
	browserCookies, err := browserSource.Cookies(ctx.CommandContext())
	if err != nil {
		return nil, "", err
	}
	return browserCookies, "browser", nil
}

type pastedCookies struct {
	spdc  string
	spkey string
	spt   string
}

func readPastedCookies(r io.Reader, out *output.Writer, interactive bool) (pastedCookies, error) {
	if interactive {
		return promptPastedCookies(out)
	}
	return parsePastedCookies(r)
}

func promptPastedCookies(out *output.Writer) (pastedCookies, error) {
	reader := bufio.NewReader(os.Stdin)
	spdc, err := readPromptCookieValue(reader, out, "sp_dc", true)
	if err != nil {
		return pastedCookies{}, err
	}
	spkey, err := readPromptCookieValue(reader, out, "sp_key", false)
	if err != nil {
		return pastedCookies{}, err
	}
	spt, err := readPromptCookieValue(reader, out, "sp_t", false)
	if err != nil {
		return pastedCookies{}, err
	}
	return pastedCookies{spdc: spdc, spkey: spkey, spt: spt}, nil
}

func readPromptCookieValue(reader *bufio.Reader, out *output.Writer, name string, required bool) (string, error) {
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}
	if out != nil {
		_, _ = fmt.Fprintf(out.Err, "Paste %s value: ", name)
	}
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value := normalizePromptCookieValue(line, name)
	if value == "" && required {
		return "", fmt.Errorf("%s required", name)
	}
	return value, nil
}

func parsePastedCookies(r io.Reader) (pastedCookies, error) {
	if r == nil {
		r = os.Stdin
	}
	scanner := bufio.NewScanner(r)
	values := pastedCookies{}
	for scanner.Scan() {
		line := scanner.Text()
		if value, ok := extractNamedCookieValue(line, "sp_dc"); ok {
			values.spdc = value
		}
		if value, ok := extractNamedCookieValue(line, "sp_key"); ok {
			values.spkey = value
		}
		if value, ok := extractNamedCookieValue(line, "sp_t"); ok {
			values.spt = value
		}
	}
	if err := scanner.Err(); err != nil {
		return pastedCookies{}, err
	}
	return values, nil
}

func normalizeCookieDomain(domain string) string {
	trimmed := strings.TrimSpace(domain)
	if trimmed == "" {
		trimmed = "spotify.com"
	}
	if strings.Contains(trimmed, "://") {
		if parsed, err := url.Parse(trimmed); err == nil && parsed.Hostname() != "" {
			trimmed = parsed.Hostname()
		}
	}
	if !strings.HasPrefix(trimmed, ".") {
		trimmed = "." + trimmed
	}
	return trimmed
}

func normalizeCookiePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/"
	}
	return trimmed
}

func normalizePromptCookieValue(value, name string) string {
	if parsed, ok := extractNamedCookieValue(value, name); ok {
		return parsed
	}
	return trimCookieValue(value)
}

func extractNamedCookieValue(value, name string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	trimmed = strings.Trim(trimmed, "\"'")
	for _, part := range strings.Split(trimmed, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, found := strings.Cut(part, "=")
		if !found {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), name) {
			return trimCookieValue(val), true
		}
	}
	return "", false
}

func trimCookieValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), "\"'")
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
