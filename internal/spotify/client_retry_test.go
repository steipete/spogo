package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type countingTokenProvider struct {
	calls int
}

func (p *countingTokenProvider) Token(context.Context) (Token, error) {
	p.calls++
	return Token{AccessToken: "token", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func TestClientRetriesOnRateLimit(t *testing.T) {
	provider := &countingTokenProvider{}
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/me/player/devices", func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"status":  http.StatusTooManyRequests,
					"message": "rate limit",
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(deviceResponse{Devices: []deviceItem{{ID: "d1", Name: "Desk"}}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewClient(Options{
		TokenProvider: provider,
		BaseURL:       srv.URL,
		HTTPClient:    srv.Client(),
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	if _, err := client.Devices(context.Background()); err != nil {
		t.Fatalf("devices: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
	if provider.calls < 2 {
		t.Fatalf("expected token refresh, got %d calls", provider.calls)
	}
}
