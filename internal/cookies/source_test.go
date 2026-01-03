package cookies

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/steipete/sweetcookie"
)

func TestBrowserSource(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		exp := time.Now().Add(time.Hour)
		return sweetcookie.Result{
			Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com", Expires: &exp}},
		}, nil
	})
	defer restore()
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
	restore := SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{}, nil
	})
	defer restore()
	src := BrowserSource{Domain: "spotify.com"}
	if _, err := src.Cookies(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSetReadCookies(t *testing.T) {
	restore := SetReadCookies(nil)
	restore()
	restore = SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{}, nil
	})
	restore()
}

func TestBrowserSourceError(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{}, errors.New("boom")
	})
	defer restore()
	src := BrowserSource{Domain: "spotify.com"}
	if _, err := src.Cookies(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestBrowserSourceDefaultDomain(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{
			Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}},
		}, nil
	})
	defer restore()
	src := BrowserSource{}
	if _, err := src.Cookies(context.Background()); err != nil {
		t.Fatalf("expected cookies")
	}
}

func TestBrowserSourceWithProfileFilter(t *testing.T) {
	restore := SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		return sweetcookie.Result{
			Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}},
		}, nil
	})
	defer restore()
	src := BrowserSource{Browser: "chrome", Profile: "Default", Domain: "spotify.com"}
	cookies, err := src.Cookies(context.Background())
	if err != nil || len(cookies) != 1 {
		t.Fatalf("expected cookie")
	}
}

func TestBrowserSourceProfileOnlyUsesDefaults(t *testing.T) {
	var got sweetcookie.Options
	restore := SetReadCookies(func(ctx context.Context, opts sweetcookie.Options) (sweetcookie.Result, error) {
		got = opts
		return sweetcookie.Result{
			Cookies: []sweetcookie.Cookie{{Name: "sp_dc", Value: "token", Domain: ".spotify.com"}},
		}, nil
	})
	defer restore()
	src := BrowserSource{Profile: "Default"}
	if _, err := src.Cookies(context.Background()); err != nil {
		t.Fatalf("expected cookies")
	}
	if got.Profiles == nil || len(got.Profiles) == 0 {
		t.Fatalf("expected default profiles map")
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
