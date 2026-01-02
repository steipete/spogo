package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePlaylistMissingUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(userProfile{})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	client, err := NewClient(Options{TokenProvider: staticTokenProvider{}, BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	if _, err := client.CreatePlaylist(context.Background(), "Name", true, false); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSearchMissingContainer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"album": map[string]any{}})
	})
	client, closeFn := newTestClient(t, handler)
	defer closeFn()
	if _, err := client.Search(context.Background(), "track", "song", 1, 0); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewClientMissingTokenProvider(t *testing.T) {
	if _, err := NewClient(Options{}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetItemErrors(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	client, closeFn := newTestClient(t, handler)
	defer closeFn()
	if _, err := client.GetTrack(context.Background(), "t1"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetAlbum(context.Background(), "a1"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetArtist(context.Background(), "ar1"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetPlaylist(context.Background(), "p1"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetShow(context.Background(), "s1"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetEpisode(context.Background(), "e1"); err == nil {
		t.Fatalf("expected error")
	}
}
