package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchTotpSecretHTTPURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string][]int{
			"42": {1, 2, 3},
		})
	}))
	defer srv.Close()

	version, secret, err := fetchTotpSecretHTTPURL(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if version != 42 {
		t.Fatalf("unexpected version: %d", version)
	}
	if len(secret) != 3 {
		t.Fatalf("unexpected secret length")
	}
}

func TestFetchTotpSecretHTTPEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string][]int{
			"7": {9, 8, 7},
		})
	}))
	defer srv.Close()

	t.Setenv(totpSecretEnv, srv.URL)
	version, secret, err := fetchTotpSecretHTTP(context.Background())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if version != 7 || len(secret) != 3 {
		t.Fatalf("unexpected secret: version=%d len=%d", version, len(secret))
	}
}

func TestFetchTotpSecretFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(path, []byte(`{"5":[1,2,3,4]}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	version, secret, err := fetchTotpSecretSource(context.Background(), path)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if version != 5 || len(secret) != 4 {
		t.Fatalf("unexpected secret: version=%d len=%d", version, len(secret))
	}
}

func TestTotpSecretSourcesDefault(t *testing.T) {
	prev := totpSecretURLs
	totpSecretURLs = []string{"one", "two"}
	t.Cleanup(func() { totpSecretURLs = prev })

	t.Setenv(totpSecretEnv, "")
	sources := totpSecretSources()
	if len(sources) != 2 {
		t.Fatalf("unexpected sources: %#v", sources)
	}
	sources[0] = "mutated"
	if totpSecretURLs[0] == "mutated" {
		t.Fatalf("expected copy")
	}
}

func TestFetchTotpSecretHTTPFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string][]int{"9": {1}})
	}))
	defer srv.Close()

	prev := totpSecretURLs
	totpSecretURLs = []string{"http://127.0.0.1:1", srv.URL}
	t.Cleanup(func() { totpSecretURLs = prev })
	t.Setenv(totpSecretEnv, "")

	version, secret, err := fetchTotpSecretHTTP(context.Background())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if version != 9 || len(secret) != 1 {
		t.Fatalf("unexpected secret")
	}
}

func TestFetchTotpSecretSourceFileScheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(path, []byte(`{"11":[1,2]}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	version, secret, err := fetchTotpSecretSource(context.Background(), "file://"+path)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if version != 11 || len(secret) != 2 {
		t.Fatalf("unexpected secret")
	}
	if _, _, err := fetchTotpSecretSource(context.Background(), ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestFetchTotpSecretHTTPAllFail(t *testing.T) {
	prev := totpSecretURLs
	totpSecretURLs = []string{"http://127.0.0.1:2"}
	t.Cleanup(func() { totpSecretURLs = prev })
	t.Setenv(totpSecretEnv, "")

	if _, _, err := fetchTotpSecretHTTP(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseTotpSecretBadJSON(t *testing.T) {
	if _, _, err := parseTotpSecret(strings.NewReader("nope")); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadTotpSecretFileEmpty(t *testing.T) {
	if _, _, err := loadTotpSecretFile(""); err == nil {
		t.Fatalf("expected error")
	}
}
