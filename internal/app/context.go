package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/cookies"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
)

type Settings struct {
	ConfigPath string
	Profile    string
	Timeout    time.Duration
	Market     string
	Language   string
	Device     string
	Engine     string
	Format     output.Format
	NoColor    bool
	Quiet      bool
	Verbose    bool
	Debug      bool
	NoInput    bool
}

type Context struct {
	Settings   Settings
	Config     *config.Config
	ConfigPath string
	Profile    config.Profile
	ProfileKey string
	Output     *output.Writer

	spotifyClient spotify.API
	commandCtx    context.Context
}

func NewContext(settings Settings) (*Context, error) {
	configPath, err := resolveConfigPath(settings.ConfigPath)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	profileKey := resolveProfileKey(cfg, settings.Profile)
	profile := applySettings(cfg.Profile(profileKey), settings)
	writer, err := newOutputWriter(settings)
	if err != nil {
		return nil, err
	}
	return &Context{
		Settings:   settings,
		Config:     cfg,
		ConfigPath: configPath,
		Profile:    profile,
		ProfileKey: profileKey,
		Output:     writer,
		commandCtx: context.Background(),
	}, nil
}

func (c *Context) Spotify() (spotify.API, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}
	if c.spotifyClient != nil {
		return c.spotifyClient, nil
	}
	source, err := c.cookieSource()
	if err != nil {
		return nil, err
	}
	client, err := c.buildSpotifyClient(source)
	if err != nil {
		return nil, err
	}
	c.spotifyClient = client
	return client, nil
}

func (c *Context) buildSpotifyClient(source cookies.Source) (spotify.API, error) {
	switch c.engine() {
	case "connect":
		return c.newConnectClient(source)
	case "web":
		webClient, err := c.newWebClient(source)
		if err != nil {
			return nil, err
		}
		client := spotify.API(webClient)
		if connectClient, connectErr := c.newConnectClient(source); connectErr == nil {
			client = spotify.NewPlaybackFallbackClient(webClient, connectClient)
		}
		return client, nil
	case "auto":
		webClient, err := c.newWebClient(source)
		if err != nil {
			return nil, err
		}
		client := spotify.API(webClient)
		if connectClient, connectErr := c.newConnectClient(source); connectErr == nil {
			client = spotify.NewAutoClient(connectClient, webClient)
		}
		return client, nil
	case "applescript":
		var fallback spotify.API
		if webClient, webErr := c.newWebClient(source); webErr == nil {
			fallback = webClient
		}
		client, err := spotify.NewAppleScriptClient(spotify.AppleScriptOptions{
			Fallback: fallback,
		})
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unknown engine %q (use auto, web, connect, or applescript)", c.engine())
	}
}

func (c *Context) newConnectClient(source cookies.Source) (*spotify.ConnectClient, error) {
	return spotify.NewConnectClient(spotify.ConnectOptions{
		Source:   source,
		Market:   c.Profile.Market,
		Language: c.Profile.Language,
		Device:   c.Profile.Device,
		Timeout:  c.Settings.Timeout,
	})
}

func (c *Context) newWebClient(source cookies.Source) (*spotify.Client, error) {
	return spotify.NewClient(spotify.Options{
		TokenProvider: spotify.CookieTokenProvider{Source: source},
		Market:        c.Profile.Market,
		Language:      c.Profile.Language,
		Device:        c.Profile.Device,
		Timeout:       c.Settings.Timeout,
	})
}

func (c *Context) engine() string {
	engine := strings.ToLower(strings.TrimSpace(c.Profile.Engine))
	if engine == "" {
		return "connect"
	}
	return engine
}

func (c *Context) SetSpotify(client spotify.API) {
	if c == nil {
		return
	}
	c.spotifyClient = client
}

func (c *Context) cookieSource() (cookies.Source, error) {
	// Use explicit cookie path from config if set
	if c.Profile.CookiePath != "" {
		return cookies.FileSource{Path: c.Profile.CookiePath}, nil
	}
	// Check if cookies exist at the default path (supports headless servers
	// where cookies were copied manually without running auth import)
	defaultPath := c.ResolveCookiePath()
	if defaultPath != "" {
		if _, err := os.Stat(defaultPath); err == nil {
			return cookies.FileSource{Path: defaultPath}, nil
		}
	}
	// Fall back to reading from browser
	browser := c.Profile.Browser
	if strings.TrimSpace(browser) == "" {
		browser = "chrome"
	}
	return cookies.BrowserSource{
		Browser: browser,
		Profile: c.Profile.BrowserProfile,
		Domain:  "spotify.com",
	}, nil
}

func (c *Context) SaveProfile(profile config.Profile) error {
	if c == nil {
		return errors.New("nil context")
	}
	if c.Config == nil {
		return errors.New("nil config")
	}
	cfg := c.Config
	cfg.SetProfile(c.ProfileKey, profile)
	cfg.DefaultProfile = c.ProfileKey
	if err := config.Save(c.ConfigPath, cfg); err != nil {
		return err
	}
	c.Profile = profile
	return nil
}

func isColorEnabled(format output.Format, noColor bool) bool {
	if format != output.FormatHuman {
		return false
	}
	if noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "dumb" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd())
}

func (c *Context) ResolveCookiePath() string {
	return config.CookiePath(c.ConfigPath, c.ProfileKey)
}

func (c *Context) EnsureTimeout() time.Duration {
	if c.Settings.Timeout > 0 {
		return c.Settings.Timeout
	}
	return 10 * time.Second
}

func (c *Context) CommandContext() context.Context {
	if c == nil || c.commandCtx == nil {
		return context.Background()
	}
	return c.commandCtx
}

func (c *Context) SetCommandContext(ctx context.Context) {
	if c == nil {
		return
	}
	if ctx == nil {
		c.commandCtx = context.Background()
		return
	}
	c.commandCtx = ctx
}

func (c *Context) ValidateProfile() error {
	if c.Profile.Market != "" && len(c.Profile.Market) != 2 {
		return fmt.Errorf("market must be 2-letter country code")
	}
	return nil
}

func resolveConfigPath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	return config.DefaultPath()
}

func resolveProfileKey(cfg *config.Config, requested string) string {
	if requested != "" {
		return requested
	}
	if cfg != nil && cfg.DefaultProfile != "" {
		return cfg.DefaultProfile
	}
	return config.DefaultProfile
}

func applySettings(profile config.Profile, settings Settings) config.Profile {
	if settings.Market != "" {
		profile.Market = settings.Market
	}
	if settings.Language != "" {
		profile.Language = settings.Language
	}
	if settings.Device != "" {
		profile.Device = settings.Device
	}
	if settings.Engine != "" {
		profile.Engine = settings.Engine
	}
	return profile
}

func newOutputWriter(settings Settings) (*output.Writer, error) {
	format := settings.Format
	if format == "" {
		format = output.FormatHuman
	}
	colorEnabled := isColorEnabled(format, settings.NoColor)
	return output.New(output.Options{
		Format: format,
		Color:  colorEnabled,
		Quiet:  settings.Quiet,
	})
}
