package app

import (
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
	Format     output.Format
	NoColor    bool
	Quiet      bool
	Verbose    bool
	Debug      bool
}

type Context struct {
	Settings   Settings
	Config     *config.Config
	ConfigPath string
	Profile    config.Profile
	ProfileKey string
	Output     *output.Writer

	spotifyClient spotify.API
}

func NewContext(settings Settings) (*Context, error) {
	configPath := settings.ConfigPath
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if configPath == "" {
		configPath, err = config.DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	profileKey := settings.Profile
	if profileKey == "" {
		profileKey = cfg.DefaultProfile
	}
	if profileKey == "" {
		profileKey = config.DefaultProfile
	}
	profile := cfg.Profile(profileKey)
	if settings.Market != "" {
		profile.Market = settings.Market
	}
	if settings.Language != "" {
		profile.Language = settings.Language
	}
	if settings.Device != "" {
		profile.Device = settings.Device
	}
	format := settings.Format
	if format == "" {
		format = output.FormatHuman
	}
	colorEnabled := isColorEnabled(format, settings.NoColor)
	writer, err := output.New(output.Options{
		Format: format,
		Color:  colorEnabled,
		Quiet:  settings.Quiet,
	})
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
	provider := spotify.CookieTokenProvider{
		Source: source,
	}
	client, err := spotify.NewClient(spotify.Options{
		TokenProvider: provider,
		Market:        c.Profile.Market,
		Language:      c.Profile.Language,
		Device:        c.Profile.Device,
		Timeout:       c.Settings.Timeout,
	})
	if err != nil {
		return nil, err
	}
	c.spotifyClient = client
	return client, nil
}

func (c *Context) SetSpotify(client spotify.API) {
	if c == nil {
		return
	}
	c.spotifyClient = client
}

func (c *Context) cookieSource() (cookies.Source, error) {
	if c.Profile.CookiePath != "" {
		return cookies.FileSource{Path: c.Profile.CookiePath}, nil
	}
	return cookies.BrowserSource{
		Browser: c.Profile.Browser,
		Profile: c.Profile.BrowserProfile,
		Domain:  "spotify.com",
	}, nil
}

func (c *Context) SaveProfile(profile config.Profile) error {
	if c == nil {
		return errors.New("nil context")
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

func (c *Context) ValidateProfile() error {
	if c.Profile.Market != "" && len(c.Profile.Market) != 2 {
		return fmt.Errorf("market must be 2-letter country code")
	}
	return nil
}
