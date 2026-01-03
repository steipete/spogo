package cookies

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/steipete/sweetcookie"
)

var readCookies = sweetcookie.Get

// SetReadCookies overrides the internal cookie reader and returns a restore func.
// Intended for tests.
func SetReadCookies(fn func(context.Context, sweetcookie.Options) (sweetcookie.Result, error)) func() {
	prev := readCookies
	if fn == nil {
		readCookies = sweetcookie.Get
	} else {
		readCookies = fn
	}
	return func() { readCookies = prev }
}

type Source interface {
	Cookies(ctx context.Context) ([]*http.Cookie, error)
}

type BrowserSource struct {
	Browser string
	Profile string
	Domain  string
}

type FileSource struct {
	Path string
}

func (s BrowserSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	domain := strings.TrimSpace(s.Domain)
	if domain == "" {
		domain = "spotify.com"
	}
	host := domain
	if strings.Contains(domain, "://") {
		if parsed, err := url.Parse(domain); err == nil && parsed.Hostname() != "" {
			host = parsed.Hostname()
		}
	}
	url := "https://" + host
	origins := []string{}
	if strings.Contains(host, "spotify.com") {
		if host != "open.spotify.com" {
			origins = append(origins, "https://open.spotify.com")
		}
		if host != "spotify.com" {
			origins = append(origins, "https://spotify.com")
		}
	}
	opts := sweetcookie.Options{
		URL:     url,
		Origins: origins,
		Mode:    sweetcookie.ModeFirst,
		Timeout: 5 * time.Second,
	}
	if s.Browser != "" {
		browser := sweetcookie.Browser(strings.ToLower(s.Browser))
		opts.Browsers = []sweetcookie.Browser{browser}
		if s.Profile != "" {
			opts.Profiles = map[sweetcookie.Browser]string{browser: s.Profile}
		}
	} else if s.Profile != "" {
		opts.Profiles = map[sweetcookie.Browser]string{}
		for _, browser := range sweetcookie.DefaultBrowsers() {
			opts.Profiles[browser] = s.Profile
		}
	}
	result, err := readCookies(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(result.Cookies) == 0 {
		return nil, errors.New("no cookies found")
	}
	ret := make([]*http.Cookie, 0, len(result.Cookies))
	for _, c := range result.Cookies {
		cookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HttpOnly: c.HTTPOnly,
		}
		if c.Expires != nil {
			cookie.Expires = *c.Expires
		}
		ret = append(ret, cookie)
	}
	return ret, nil
}

func (s FileSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	_ = ctx
	return Read(s.Path)
}
