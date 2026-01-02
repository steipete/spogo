package cookies

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
)

var readCookies = kooky.ReadCookies

// SetReadCookies overrides the internal cookie reader and returns a restore func.
// Intended for tests.
func SetReadCookies(fn func(context.Context, ...kooky.Filter) (kooky.Cookies, error)) func() {
	prev := readCookies
	if fn == nil {
		readCookies = kooky.ReadCookies
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
	filters := []kooky.Filter{kooky.DomainHasSuffix(domain)}
	if s.Browser != "" || s.Profile != "" {
		filters = append(filters, kooky.FilterFunc(func(cookie *kooky.Cookie) bool {
			if cookie == nil || cookie.Browser == nil {
				return false
			}
			if s.Browser != "" && cookie.Browser.Browser() != s.Browser {
				return false
			}
			if s.Profile != "" && cookie.Browser.Profile() != s.Profile {
				return false
			}
			return true
		}))
	}
	cookies, err := readCookies(ctx, filters...)
	if err != nil {
		return nil, err
	}
	if len(cookies) == 0 {
		return nil, errors.New("no cookies found")
	}
	ret := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		if c == nil {
			continue
		}
		cookie := c.Cookie
		ret = append(ret, &cookie)
	}
	return ret, nil
}

func (s FileSource) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	_ = ctx
	return Read(s.Path)
}
