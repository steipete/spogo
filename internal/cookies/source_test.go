package cookies

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/browserutils/kooky"
)

func TestBrowserSource(t *testing.T) {
	orig := readCookies
	defer func() { readCookies = orig }()
	readCookies = func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		cookie := &kooky.Cookie{Cookie: http.Cookie{Name: "sp_dc", Value: "token", Domain: ".spotify.com", Expires: time.Now().Add(time.Hour)}}
		return kooky.Cookies{cookie}, nil
	}
	src := BrowserSource{Browser: "chrome", Profile: "Default", Domain: "spotify.com"}
	cookies, err := src.Cookies(context.Background())
	if err != nil {
		t.Fatalf("cookies: %v", err)
	}
	if len(cookies) != 1 || cookies[0].Name != "sp_dc" {
		t.Fatalf("unexpected cookies: %#v", cookies)
	}
}

func TestBrowserSourceNoCookies(t *testing.T) {
	orig := readCookies
	defer func() { readCookies = orig }()
	readCookies = func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{}, nil
	}
	src := BrowserSource{Domain: "spotify.com"}
	if _, err := src.Cookies(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSetReadCookies(t *testing.T) {
	restore := SetReadCookies(nil)
	restore()
	restore = SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{}, nil
	})
	restore()
}

func TestBrowserSourceError(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		return nil, errors.New("boom")
	})
	defer restore()
	src := BrowserSource{Domain: "spotify.com"}
	if _, err := src.Cookies(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestBrowserSourceDefaultDomain(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		cookie := &kooky.Cookie{Cookie: http.Cookie{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}}
		return kooky.Cookies{cookie}, nil
	})
	defer restore()
	src := BrowserSource{}
	if _, err := src.Cookies(context.Background()); err != nil {
		t.Fatalf("expected cookies")
	}
}

type dummyBrowserInfo struct {
	browser string
	profile string
}

func (d dummyBrowserInfo) Browser() string         { return d.browser }
func (d dummyBrowserInfo) Profile() string         { return d.profile }
func (d dummyBrowserInfo) IsDefaultProfile() bool  { return false }
func (d dummyBrowserInfo) FilePath() string        { return "" }

func TestBrowserSourceWithProfileFilter(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, filters ...kooky.Filter) (kooky.Cookies, error) {
		cookie := &kooky.Cookie{
			Cookie:  http.Cookie{Name: "sp_dc", Value: "token", Domain: ".spotify.com"},
			Browser: dummyBrowserInfo{browser: "chrome", profile: "Default"},
		}
		if !kooky.FilterCookie(ctx, cookie, filters...) {
			return kooky.Cookies{}, nil
		}
		return kooky.Cookies{cookie}, nil
	})
	defer restore()
	src := BrowserSource{Browser: "chrome", Profile: "Default", Domain: "spotify.com"}
	cookies, err := src.Cookies(context.Background())
	if err != nil || len(cookies) != 1 {
		t.Fatalf("expected cookie")
	}
}

func TestFileSource(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/cookies.json"
	if err := Write(path, []*http.Cookie{{Name: "sp_dc", Value: "token"}}); err != nil {
		t.Fatalf("write: %v", err)
	}
	src := FileSource{Path: path}
	cookies, err := src.Cookies(context.Background())
	if err != nil {
		t.Fatalf("cookies: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie")
	}
}

func TestFileSourceError(t *testing.T) {
	src := FileSource{Path: "/nope/missing.json"}
	if _, err := src.Cookies(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}
