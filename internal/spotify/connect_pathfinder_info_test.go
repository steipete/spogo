package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConnectLibraryV3Helpers(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Query().Get("operationName") {
		case "libraryV3":
			query := req.URL.Query()
			variables := query.Get("variables")
			switch {
			case strings.Contains(variables, `"Playlists"`):
				return jsonResponse(http.StatusOK, map[string]any{
					"data": map[string]any{"me": map[string]any{"libraryV3": map[string]any{
						"totalCount": 1,
						"items": []any{map[string]any{"item": map[string]any{"data": map[string]any{
							"uri":         "spotify:playlist:p1",
							"name":        "Playlist",
							"ownerV2":     map[string]any{"data": map[string]any{"name": "Owner"}},
							"totalTracks": 12,
						}}}},
					}}},
				}), nil
			case strings.Contains(variables, `"Albums"`):
				return jsonResponse(http.StatusOK, map[string]any{
					"data": map[string]any{"me": map[string]any{"libraryV3": map[string]any{
						"totalCount": 1,
						"items": []any{map[string]any{"item": map[string]any{"data": map[string]any{
							"uri":  "spotify:album:a1",
							"name": "Album",
						}}}},
					}}},
				}), nil
			}
		case "fetchLibraryTracks":
			return jsonResponse(http.StatusOK, map[string]any{
				"data": map[string]any{"me": map[string]any{"library": map[string]any{"tracks": map[string]any{
					"totalCount": 1,
					"items": []any{map[string]any{"track": map[string]any{
						"_uri": "spotify:track:t1",
						"data": map[string]any{
							"name": "Song",
						},
					}}},
				}}}},
			}), nil
		case "fetchPlaylist":
			return jsonResponse(http.StatusOK, map[string]any{
				"data": map[string]any{"playlistV2": map[string]any{"content": map[string]any{
					"totalCount": 1,
					"items": []any{map[string]any{"itemV2": map[string]any{"data": map[string]any{
						"track": map[string]any{"uri": "spotify:track:t1", "name": "Song"},
					}}}},
				}}},
			}), nil
		}
		return textResponse(http.StatusNotFound, "missing"), nil
	})
	client := newConnectClientForTests(transport)
	for _, op := range []string{"libraryV3", "fetchPlaylist", "fetchLibraryTracks"} {
		client.hashes.hashes[op] = "hash"
	}

	playlists, total, err := client.playlists(context.Background(), 10, 0)
	if err != nil || total != 1 || len(playlists) != 1 || playlists[0].ID != "p1" {
		t.Fatalf("playlists: items=%#v total=%d err=%v", playlists, total, err)
	}
	tracks, total, err := client.playlistTracks(context.Background(), "p1", 10, 0)
	if err != nil || total != 1 || len(tracks) != 1 || tracks[0].ID != "t1" {
		t.Fatalf("playlist tracks: items=%#v total=%d err=%v", tracks, total, err)
	}
	libraryTracks, total, err := client.libraryTracks(context.Background(), 10, 0)
	if err != nil || total != 1 || len(libraryTracks) != 1 || libraryTracks[0].ID != "t1" {
		t.Fatalf("library tracks: items=%#v total=%d err=%v", libraryTracks, total, err)
	}
	albums, total, err := client.libraryAlbums(context.Background(), 10, 0)
	if err != nil || total != 1 || len(albums) != 1 || albums[0].ID != "a1" {
		t.Fatalf("library albums: items=%#v total=%d err=%v", albums, total, err)
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
