package spotify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/steipete/spogo/internal/cookies"
)

type connectSession struct {
	source cookies.Source
	client *http.Client
	cache  *connectCacheStore

	mu          sync.Mutex
	cacheLoaded bool

	token        Token
	clientToken  string
	clientTokenT time.Time
	clientID     string
	clientVer    string
	connectVer   string
	deviceID     string

	connectDeviceID string
	connectionID    string
	registeredAt    time.Time
}

type connectAuth struct {
	AccessToken    string
	ClientToken    string
	ClientVersion  string
	ConnectVersion string
	DeviceID       string
}

func (s *connectSession) auth(ctx context.Context) (connectAuth, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadCacheLocked()
	if err := s.ensureTokenLocked(ctx); err != nil {
		return connectAuth{}, err
	}
	if err := s.ensureAppConfigLocked(ctx); err != nil {
		return connectAuth{}, err
	}
	if err := s.ensureClientTokenLocked(ctx); err != nil {
		return connectAuth{}, err
	}
	return connectAuth{
		AccessToken:    s.token.AccessToken,
		ClientToken:    s.clientToken,
		ClientVersion:  s.clientVer,
		ConnectVersion: s.connectVer,
		DeviceID:       s.deviceID,
	}, nil
}

func (s *connectSession) loadCacheLocked() {
	if s == nil || s.cacheLoaded {
		return
	}
	s.cacheLoaded = true
	if s.cache == nil {
		return
	}
	cached, err := s.cache.load()
	if err != nil {
		return
	}
	if cached.AccessToken != "" && cached.AccessTokenExpiresUnix > 0 {
		s.token = Token{
			AccessToken: cached.AccessToken,
			ExpiresAt:   time.Unix(cached.AccessTokenExpiresUnix, 0),
			Anonymous:   cached.Anonymous,
			ClientID:    cached.ClientID,
		}
	}
	if cached.ClientID != "" {
		s.clientID = cached.ClientID
	}
	if cached.ClientToken != "" {
		s.clientToken = cached.ClientToken
	}
	if cached.ClientTokenExpiresUnix > 0 {
		s.clientTokenT = time.Unix(cached.ClientTokenExpiresUnix, 0)
	}
	if cached.ClientVersion != "" {
		s.clientVer = cached.ClientVersion
	}
	if cached.DeviceID != "" {
		s.deviceID = cached.DeviceID
	}
	if cached.ConnectDeviceID != "" {
		s.connectDeviceID = cached.ConnectDeviceID
	}
}

func (s *connectSession) saveCacheLocked() {
	if s == nil || s.cache == nil {
		return
	}
	token := s.token
	clientID := s.clientID
	if clientID == "" {
		clientID = token.ClientID
	}
	clientToken := s.clientToken
	clientTokenT := s.clientTokenT
	clientVer := s.clientVer
	deviceID := s.deviceID
	connectDeviceID := s.connectDeviceID
	_ = s.cache.update(func(cached *connectCache) {
		cached.AccessToken = token.AccessToken
		cached.AccessTokenExpiresUnix = unixOrZero(token.ExpiresAt)
		cached.Anonymous = token.Anonymous
		cached.ClientID = clientID
		cached.ClientToken = clientToken
		cached.ClientTokenExpiresUnix = unixOrZero(clientTokenT)
		cached.ClientVersion = clientVer
		cached.ConnectVersion = ""
		cached.DeviceID = deviceID
		cached.ConnectDeviceID = connectDeviceID
		cached.ActiveDeviceID = ""
		cached.OriginDeviceID = ""
		cached.RouteUnix = 0
	})
}

func (s *connectSession) ensureTokenLocked(ctx context.Context) error {
	if s.token.AccessToken != "" && time.Until(s.token.ExpiresAt) > time.Minute {
		return nil
	}
	provider := CookieTokenProvider{Source: s.source, Client: s.client}
	token, err := provider.Token(ctx)
	if err != nil {
		return err
	}
	s.token = token
	if token.ClientID != "" {
		s.clientID = token.ClientID
	}
	s.saveCacheLocked()
	return nil
}

func (s *connectSession) ensureAppConfigLocked(ctx context.Context) error {
	s.connectVer = connectClientVersion()
	if s.clientVer != "" && s.deviceID != "" {
		return nil
	}
	cookiesList, err := s.source.Cookies(ctx)
	if err != nil {
		return err
	}
	deviceID := ""
	for _, cookie := range cookiesList {
		if cookie.Name == "sp_t" {
			deviceID = cookie.Value
			break
		}
	}
	if deviceID == "" {
		return errors.New("missing sp_t cookie (run `spogo auth paste` and include sp_t from DevTools)")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	baseURL, _ := url.Parse("https://open.spotify.com/")
	jar.SetCookies(baseURL, cookiesList)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://open.spotify.com/", nil)
	if err != nil {
		return err
	}
	applyRequestHeaders(req, requestHeaders{})
	client := *s.client
	client.Jar = jar
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiErrorFromResponse(resp)
	}
	body, err := readAll(resp)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`<script id="appServerConfig" type="text/plain">([^<]+)</script>`)
	match := re.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return errors.New("missing appServerConfig")
	}
	raw, err := base64.StdEncoding.DecodeString(match[1])
	if err != nil {
		return err
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	clientVer, _ := payload["clientVersion"].(string)
	if clientVer == "" {
		return errors.New("missing clientVersion")
	}
	if idx := strings.Index(clientVer, ".g"); idx > 0 {
		clientVer = clientVer[:idx]
	}
	s.clientVer = clientVer
	s.deviceID = deviceID
	s.saveCacheLocked()
	return nil
}

func connectClientVersion() string {
	if override := strings.TrimSpace(os.Getenv("SPOGO_CONNECT_VERSION")); override != "" {
		return override
	}
	return "harmony:4.43.2-a61ecaf5"
}

func (s *connectSession) ensureClientTokenLocked(ctx context.Context) error {
	if s.clientToken != "" && time.Until(s.clientTokenT) > time.Minute {
		return nil
	}
	if s.clientID == "" {
		return errors.New("missing client id")
	}
	osName, osVersion := runtimeOS()
	payload := map[string]any{
		"client_data": map[string]any{
			"client_version": s.clientVer,
			"client_id":      s.clientID,
			"js_sdk_data": map[string]any{
				"device_brand": "unknown",
				"device_model": "unknown",
				"os":           osName,
				"os_version":   osVersion,
				"device_id":    s.deviceID,
				"device_type":  "computer",
			},
		},
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://clienttoken.spotify.com/v1/clienttoken", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	applyRequestHeaders(req, requestHeaders{
		ContentType: "application/json",
		Accept:      "application/json",
	})
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiErrorFromResponse(resp)
	}
	var tokenPayload struct {
		ResponseType string `json:"response_type"`
		GrantedToken struct {
			Token   string `json:"token"`
			Expires int    `json:"expires_in"`
		} `json:"granted_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenPayload); err != nil {
		return err
	}
	if tokenPayload.GrantedToken.Token == "" {
		return errors.New("missing client token")
	}
	s.clientToken = tokenPayload.GrantedToken.Token
	if tokenPayload.GrantedToken.Expires > 0 {
		s.clientTokenT = time.Now().Add(time.Duration(tokenPayload.GrantedToken.Expires) * time.Second)
	} else {
		s.clientTokenT = time.Now().Add(30 * time.Minute)
	}
	s.saveCacheLocked()
	return nil
}

func runtimeOS() (string, string) {
	switch runtime.GOOS {
	case "darwin":
		return "macos", "unknown"
	case "windows":
		return "windows", "unknown"
	default:
		return "linux", "unknown"
	}
}

func readAll(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("empty response")
	}
	return io.ReadAll(resp.Body)
}
