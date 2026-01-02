package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/steipete/spogo/internal/cookies"
)

type Token struct {
	AccessToken string
	ExpiresAt   time.Time
	Anonymous   bool
}

type TokenProvider interface {
	Token(ctx context.Context) (Token, error)
}

type CookieTokenProvider struct {
	Source  cookies.Source
	BaseURL string
	Client  *http.Client
}

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
	IsAnonymous bool   `json:"isAnonymous"`
}

func (p CookieTokenProvider) Token(ctx context.Context) (Token, error) {
	if p.Source == nil {
		return Token{}, errors.New("cookie source required")
	}
	cookiesList, err := p.Source.Cookies(ctx)
	if err != nil {
		return Token{}, err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return Token{}, err
	}
	base := p.BaseURL
	if base == "" {
		base = "https://open.spotify.com/"
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	baseURL, _ := url.Parse(base)
	jar.SetCookies(baseURL, cookiesList)
	client := p.Client
	if client == nil {
		client = &http.Client{Jar: jar}
	} else {
		client.Jar = jar
	}
	reqURL := base + "get_access_token?reason=transport&productType=web_player"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("User-Agent", defaultUserAgent())
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Token{}, apiErrorFromResponse(resp)
	}
	var payload tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Token{}, err
	}
	if payload.AccessToken == "" {
		return Token{}, errors.New("missing access token")
	}
	return Token{
		AccessToken: payload.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
		Anonymous:   payload.IsAnonymous,
	}, nil
}
