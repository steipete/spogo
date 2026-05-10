package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type cookieSourceStub struct {
	cookies []*http.Cookie
	err     error
}

func (s cookieSourceStub) Cookies(ctx context.Context) ([]*http.Cookie, error) {
	return s.cookies, s.err
}

func TestConnectSessionAuth(t *testing.T) {
	restore := SetTotpSecretFetcher(func(ctx context.Context) (int, []byte, error) {
		return 1, []byte{1, 2, 3, 4}, nil
	})
	t.Cleanup(restore)

	cookies := []*http.Cookie{
		{Name: "sp_dc", Value: "token", Domain: ".spotify.com"},
		{Name: "sp_t", Value: "device", Domain: ".spotify.com"},
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Host == "open.spotify.com" && req.URL.Path == "/api/token":
			return jsonResponse(http.StatusOK, tokenResponse{
				AccessToken: "access",
				ExpiresIn:   3600,
				ClientID:    "client",
			}), nil
		case req.URL.Host == "open.spotify.com" && req.URL.Path == "/":
			raw, _ := json.Marshal(map[string]any{"clientVersion": "1.2.3"})
			encoded := base64.StdEncoding.EncodeToString(raw)
			html := fmt.Sprintf(`<script id="appServerConfig" type="text/plain">%s</script>`, encoded)
			return textResponse(http.StatusOK, html), nil
		case req.URL.Host == "clienttoken.spotify.com":
			return jsonResponse(http.StatusOK, map[string]any{
				"response_type": "OK",
				"granted_token": map[string]any{
					"token":      "client-token",
					"expires_in": 600,
				},
			}), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	client := &http.Client{Transport: transport}
	session := &connectSession{source: cookieSourceStub{cookies: cookies}, client: client}
	auth, err := session.auth(context.Background())
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	if auth.AccessToken != "access" || auth.ClientToken != "client-token" || auth.ClientVersion != "1.2.3" {
		t.Fatalf("unexpected auth: %#v", auth)
	}
	if auth.DeviceID != "device" {
		t.Fatalf("unexpected device id: %s", auth.DeviceID)
	}
}

func TestConnectSessionAuthPersistsCache(t *testing.T) {
	restore := SetTotpSecretFetcher(func(ctx context.Context) (int, []byte, error) {
		return 1, []byte{1, 2, 3, 4}, nil
	})
	t.Cleanup(restore)

	cache := newConnectCacheStore(filepath.Join(t.TempDir(), "connect.json"))
	cookies := []*http.Cookie{
		{Name: "sp_dc", Value: "token", Domain: ".spotify.com"},
		{Name: "sp_t", Value: "device", Domain: ".spotify.com"},
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Host == "open.spotify.com" && req.URL.Path == "/api/token":
			return jsonResponse(http.StatusOK, tokenResponse{
				AccessToken: "access",
				ExpiresIn:   3600,
				ClientID:    "client",
			}), nil
		case req.URL.Host == "open.spotify.com" && req.URL.Path == "/":
			raw, _ := json.Marshal(map[string]any{"clientVersion": "1.2.3"})
			return textResponse(http.StatusOK, fmt.Sprintf(`<script id="appServerConfig" type="text/plain">%s</script>`, base64.StdEncoding.EncodeToString(raw))), nil
		case req.URL.Host == "clienttoken.spotify.com":
			return jsonResponse(http.StatusOK, map[string]any{
				"granted_token": map[string]any{"token": "client-token", "expires_in": 600},
			}), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	first := &connectSession{source: cookieSourceStub{cookies: cookies}, client: &http.Client{Transport: transport}, cache: cache}
	if _, err := first.auth(context.Background()); err != nil {
		t.Fatalf("auth: %v", err)
	}

	second := &connectSession{
		client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected network request: %s", req.URL)
			return textResponse(http.StatusInternalServerError, "unexpected"), nil
		})},
		cache: cache,
	}
	auth, err := second.auth(context.Background())
	if err != nil {
		t.Fatalf("cached auth: %v", err)
	}
	if auth.AccessToken != "access" || auth.ClientToken != "client-token" || auth.ClientVersion != "1.2.3" || auth.DeviceID != "device" {
		t.Fatalf("unexpected cached auth: %#v", auth)
	}
}

func TestConnectSessionAuthRecomputesConnectVersionOverrideFromCache(t *testing.T) {
	oldOverride, hadOverride := os.LookupEnv("SPOGO_CONNECT_VERSION")
	t.Cleanup(func() {
		if hadOverride {
			_ = os.Setenv("SPOGO_CONNECT_VERSION", oldOverride)
			return
		}
		_ = os.Unsetenv("SPOGO_CONNECT_VERSION")
	})

	cache := newConnectCacheStore(filepath.Join(t.TempDir(), "connect.json"))
	if err := cache.update(func(cached *connectCache) {
		cached.AccessToken = "access"
		cached.AccessTokenExpiresUnix = time.Now().Add(time.Hour).Unix()
		cached.ClientID = "client"
		cached.ClientToken = "client-token"
		cached.ClientTokenExpiresUnix = time.Now().Add(time.Hour).Unix()
		cached.ClientVersion = "1.2.3"
		cached.ConnectVersion = "cached-version"
		cached.DeviceID = "device"
	}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
	newCachedSession := func() *connectSession {
		return &connectSession{
			client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				t.Fatalf("unexpected network request: %s", req.URL)
				return textResponse(http.StatusInternalServerError, "unexpected"), nil
			})},
			cache: cache,
		}
	}

	_ = os.Setenv("SPOGO_CONNECT_VERSION", "override-one")
	auth, err := newCachedSession().auth(context.Background())
	if err != nil {
		t.Fatalf("cached auth with override: %v", err)
	}
	if auth.ConnectVersion != "override-one" {
		t.Fatalf("expected override-one, got %q", auth.ConnectVersion)
	}

	_ = os.Setenv("SPOGO_CONNECT_VERSION", "override-two")
	auth, err = newCachedSession().auth(context.Background())
	if err != nil {
		t.Fatalf("cached auth with changed override: %v", err)
	}
	if auth.ConnectVersion != "override-two" {
		t.Fatalf("expected override-two, got %q", auth.ConnectVersion)
	}

	_ = os.Unsetenv("SPOGO_CONNECT_VERSION")
	auth, err = newCachedSession().auth(context.Background())
	if err != nil {
		t.Fatalf("cached auth without override: %v", err)
	}
	if auth.ConnectVersion != connectClientVersion() {
		t.Fatalf("expected default connect version, got %q", auth.ConnectVersion)
	}
}

func TestConnectSessionMissingDeviceCookie(t *testing.T) {
	session := &connectSession{
		source: cookieSourceStub{cookies: []*http.Cookie{{Name: "sp_dc", Value: "token"}}},
		client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return textResponse(http.StatusOK, ""), nil
		})},
	}
	if err := session.ensureAppConfigLocked(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadAllNil(t *testing.T) {
	if _, err := readAll(nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRuntimeOS(t *testing.T) {
	name, version := runtimeOS()
	if name == "" || version == "" {
		t.Fatalf("expected runtime values")
	}
}

func TestEnsureClientTokenMissingID(t *testing.T) {
	session := &connectSession{
		client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return textResponse(http.StatusOK, ""), nil
		})},
	}
	if err := session.ensureClientTokenLocked(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestEnsureTokenLockedCached(t *testing.T) {
	session := &connectSession{
		token: Token{AccessToken: "access", ExpiresAt: time.Now().Add(time.Hour)},
	}
	if err := session.ensureTokenLocked(context.Background()); err != nil {
		t.Fatalf("expected cached token")
	}
}

func TestEnsureAppConfigLockedCached(t *testing.T) {
	session := &connectSession{
		clientVer: "1.0.0",
		deviceID:  "device",
	}
	if err := session.ensureAppConfigLocked(context.Background()); err != nil {
		t.Fatalf("expected cached config")
	}
}

func TestEnsureClientTokenLockedCached(t *testing.T) {
	session := &connectSession{
		clientToken:  "token",
		clientTokenT: time.Now().Add(time.Hour),
	}
	if err := session.ensureClientTokenLocked(context.Background()); err != nil {
		t.Fatalf("expected cached token")
	}
}

func TestEnsureClientTokenNoExpiry(t *testing.T) {
	session := &connectSession{
		client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, map[string]any{
				"granted_token": map[string]any{
					"token": "ct",
				},
			}), nil
		})},
		clientID:  "client",
		clientVer: "1.0.0",
		deviceID:  "device",
	}
	if err := session.ensureClientTokenLocked(context.Background()); err != nil {
		t.Fatalf("expected token")
	}
	if session.clientToken != "ct" {
		t.Fatalf("unexpected token")
	}
	if time.Until(session.clientTokenT) < 20*time.Minute {
		t.Fatalf("unexpected expiry")
	}
}

func TestConnectSessionAuthError(t *testing.T) {
	session := &connectSession{
		source: cookieSourceStub{err: context.Canceled},
		client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return textResponse(http.StatusOK, ""), nil
		})},
	}
	if _, err := session.auth(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestEnsureAppConfigLockedMissingConfig(t *testing.T) {
	session := &connectSession{
		source: cookieSourceStub{cookies: []*http.Cookie{{Name: "sp_t", Value: "device"}}},
		client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return textResponse(http.StatusOK, "<html></html>"), nil
		})},
	}
	if err := session.ensureAppConfigLocked(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestConnectClientVersionOverride(t *testing.T) {
	t.Setenv("SPOGO_CONNECT_VERSION", "custom")
	if connectClientVersion() != "custom" {
		t.Fatalf("expected override")
	}
}
