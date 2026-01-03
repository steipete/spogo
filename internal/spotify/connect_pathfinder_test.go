package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newConnectClientForTests(transport http.RoundTripper) *ConnectClient {
	client := &http.Client{Transport: transport}
	session := &connectSession{
		client:       client,
		token:        Token{AccessToken: "access", ExpiresAt: time.Now().Add(time.Hour), ClientID: "client"},
		clientToken:  "client-token",
		clientTokenT: time.Now().Add(time.Hour),
		clientVer:    "1.0.0",
		deviceID:     "device",
	}
	hashes := &hashResolver{client: client, session: session, hashes: map[string]string{}}
	return &ConnectClient{client: client, session: session, hashes: hashes}
}

func TestPathfinderSearch(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		payload := map[string]any{
			"data": map[string]any{
				"searchV2": map[string]any{
					"tracksV2": map[string]any{
						"totalCount": 1,
						"items": []any{
							map[string]any{
								"uri":  "spotify:track:abc",
								"name": "Song",
								"artists": []any{
									map[string]any{"name": "Artist"},
								},
							},
						},
					},
				},
			},
		}
		return jsonResponse(http.StatusOK, payload), nil
	})
	client := newConnectClientForTests(transport)
	client.hashes.hashes["searchDesktop"] = "hash"
	result, err := client.Search(context.Background(), "track", "song", 1, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "Song" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestPathfinderError(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		payload := map[string]any{
			"errors": []any{map[string]any{"message": "bad"}},
		}
		return jsonResponse(http.StatusOK, payload), nil
	})
	client := newConnectClientForTests(transport)
	client.hashes.hashes["searchDesktop"] = "hash"
	_, err := client.Search(context.Background(), "track", "song", 1, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPathfinderErrorMissingMessage(t *testing.T) {
	err := pathfinderError(map[string]any{"errors": []any{map[string]any{}}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "pathfinder error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPathfinderEmptyQuery(t *testing.T) {
	client := newConnectClientForTests(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, map[string]any{}), nil
	}))
	if _, err := client.Search(context.Background(), "track", "  ", 1, 0); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInfoByOperationMissing(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, map[string]any{"data": map[string]any{}}), nil
	})
	client := newConnectClientForTests(transport)
	client.hashes.hashes["getTrack"] = "hash"
	if _, err := client.infoByOperation(context.Background(), "getTrack", map[string]any{}, "track"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPathfinderErrorEmpty(t *testing.T) {
	if err := pathfinderError(map[string]any{"errors": []any{}}); err != nil {
		t.Fatalf("unexpected error")
	}
	if err := pathfinderError(map[string]any{"errors": "bad"}); err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestGraphQLNilVariables(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, map[string]any{"data": map[string]any{}}), nil
	})
	client := newConnectClientForTests(transport)
	client.language = "en-US"
	client.hashes.hashes["searchDesktop"] = "hash"
	if _, err := client.graphQL(context.Background(), "searchDesktop", nil); err != nil {
		t.Fatalf("graphQL: %v", err)
	}
}

func TestGraphQLHTTPError(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return textResponse(http.StatusInternalServerError, "fail"), nil
	})
	client := newConnectClientForTests(transport)
	client.hashes.hashes["searchDesktop"] = "hash"
	if _, err := client.graphQL(context.Background(), "searchDesktop", map[string]any{}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPathfinderFallbackToWeb(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "api-partner.spotify.com" {
			return textResponse(http.StatusInternalServerError, "fail"), nil
		}
		payload := map[string]any{
			"track": map[string]any{
				"items": []map[string]any{{
					"id":   "t1",
					"uri":  "spotify:track:t1",
					"name": "Song",
				}},
				"limit":  1,
				"offset": 0,
				"total":  1,
			},
		}
		return jsonResponse(http.StatusOK, payload), nil
	})
	client := newConnectClientForTests(transport)
	client.hashes.hashes["searchDesktop"] = "hash"
	client.searchURL = "https://search.local/search"

	result, err := client.Search(context.Background(), "track", "song", 1, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "t1" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSearchViaWebAPIDefaultClient(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		payload := map[string]any{
			"track": map[string]any{
				"items": []map[string]any{{
					"id":   "t1",
					"uri":  "spotify:track:t1",
					"name": "Song",
				}},
				"limit":  1,
				"offset": 0,
				"total":  1,
			},
		}
		return jsonResponse(http.StatusOK, payload), nil
	})
	client := newConnectClientForTests(transport)
	client.searchURL = ""
	client.searchClient = nil

	result, err := client.searchViaWebAPI(context.Background(), "track", "song", 1, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "t1" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSearchViaWebAPIMissingKind(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		payload := map[string]any{"album": map[string]any{}}
		return jsonResponse(http.StatusOK, payload), nil
	})
	client := newConnectClientForTests(transport)
	client.searchURL = "https://search.local/search"

	if _, err := client.searchViaWebAPI(context.Background(), "track", "song", 1, 0); err == nil {
		t.Fatalf("expected error")
	}
}

func TestConnectInfoFallbackToWeb(t *testing.T) {
	infoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tracks/t1":
			_ = json.NewEncoder(w).Encode(trackItem{ID: "t1", URI: "spotify:track:t1", Name: "Song", Artists: []artistRef{{Name: "Artist"}}})
		case "/albums/a1":
			_ = json.NewEncoder(w).Encode(albumItem{ID: "a1", URI: "spotify:album:a1", Name: "Album", Artists: []artistRef{{Name: "Artist"}}})
		case "/artists/ar1":
			_ = json.NewEncoder(w).Encode(artistItem{ID: "ar1", URI: "spotify:artist:ar1", Name: "Artist"})
		case "/playlists/p1":
			_ = json.NewEncoder(w).Encode(playlistItem{ID: "p1", URI: "spotify:playlist:p1", Name: "Playlist"})
		case "/shows/s1":
			_ = json.NewEncoder(w).Encode(showItem{ID: "s1", URI: "spotify:show:s1", Name: "Show"})
		case "/episodes/e1":
			_ = json.NewEncoder(w).Encode(episodeItem{ID: "e1", URI: "spotify:episode:e1", Name: "Episode"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(infoServer.Close)

	webClient, err := NewClient(Options{
		TokenProvider: staticTokenProvider{},
		BaseURL:       infoServer.URL,
		HTTPClient:    infoServer.Client(),
	})
	if err != nil {
		t.Fatalf("web client: %v", err)
	}

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return textResponse(http.StatusInternalServerError, "fail"), nil
	})
	client := newConnectClientForTests(transport)
	client.web = webClient
	client.hashes = &hashResolver{client: &http.Client{Transport: transport}, session: client.session, hashes: map[string]string{}}

	if item, err := client.GetTrack(context.Background(), "t1"); err != nil || item.ID != "t1" {
		t.Fatalf("track: %#v err=%v", item, err)
	}
	if item, err := client.GetAlbum(context.Background(), "a1"); err != nil || item.ID != "a1" {
		t.Fatalf("album: %#v err=%v", item, err)
	}
	if item, err := client.GetArtist(context.Background(), "ar1"); err != nil || item.ID != "ar1" {
		t.Fatalf("artist: %#v err=%v", item, err)
	}
	if item, err := client.GetPlaylist(context.Background(), "p1"); err != nil || item.ID != "p1" {
		t.Fatalf("playlist: %#v err=%v", item, err)
	}
	if item, err := client.GetShow(context.Background(), "s1"); err != nil || item.ID != "s1" {
		t.Fatalf("show: %#v err=%v", item, err)
	}
	if item, err := client.GetEpisode(context.Background(), "e1"); err != nil || item.ID != "e1" {
		t.Fatalf("episode: %#v err=%v", item, err)
	}
}
