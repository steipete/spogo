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

	mu sync.Mutex

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
	return nil
}

func (s *connectSession) ensureAppConfigLocked(ctx context.Context) error {
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
		return errors.New("missing sp_t cookie")
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
	req.Header.Set("User-Agent", defaultUserAgent())
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
	s.connectVer = connectClientVersion()
	s.deviceID = deviceID
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent())
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
