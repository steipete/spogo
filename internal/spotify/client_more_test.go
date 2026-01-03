package spotify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tracks/t1", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(trackItem{ID: "t1", URI: "spotify:track:t1", Name: "Track", Album: albumRef{Name: "Album"}, Artists: []artistRef{{Name: "Artist"}}})
	})
	mux.HandleFunc("/albums/a1", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(albumItem{ID: "a1", URI: "spotify:album:a1", Name: "Album", Artists: []artistRef{{Name: "Artist"}}})
	})
	mux.HandleFunc("/artists/ar1", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(artistItem{ID: "ar1", URI: "spotify:artist:ar1", Name: "Artist"})
	})
	mux.HandleFunc("/artists/ar1/top-tracks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(artistTopTracksResponse{
			Tracks: []trackItem{{ID: "t1", URI: "spotify:track:t1", Name: "Track", Album: albumRef{Name: "Album"}}},
		})
	})
	mux.HandleFunc("/playlists/p1", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(playlistItem{ID: "p1", URI: "spotify:playlist:p1", Name: "Playlist"})
	})
	mux.HandleFunc("/shows/s1", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(showItem{ID: "s1", URI: "spotify:show:s1", Name: "Show"})
	})
	mux.HandleFunc("/episodes/e1", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(episodeItem{ID: "e1", URI: "spotify:episode:e1", Name: "Episode"})
	})
	mux.HandleFunc("/me/player/devices", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(deviceResponse{Devices: []deviceItem{{ID: "d1", Name: "Desk", Type: "speaker", Volume: 30}}})
	})
	mux.HandleFunc("/me/player", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(playbackResponse{IsPlaying: true})
	})
	mux.HandleFunc("/me/player/play", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/me/player/pause", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/me/player/next", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/me/player/previous", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/me/player/queue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(queueResponse{})
	})
	mux.HandleFunc("/me/tracks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(libraryResponse{Items: []struct {
			Track trackItem `json:"track"`
			Album albumItem `json:"album"`
		}{{Track: trackItem{ID: "t1", Name: "Track"}}}, Total: 1})
	})
	mux.HandleFunc("/me/albums", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(libraryResponse{Items: []struct {
			Track trackItem `json:"track"`
			Album albumItem `json:"album"`
		}{{Album: albumItem{ID: "a1", Name: "Album"}}}, Total: 1})
	})
	mux.HandleFunc("/me/following", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(followedArtistsResponse{Artists: artistsContainer{Items: []artistItem{{ID: "ar1", Name: "Artist"}}, Total: 1}})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/me/playlists", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(playlistListResponse{Items: []playlistItem{{ID: "p1", Name: "Playlist"}}, Total: 1})
	})
	mux.HandleFunc("/playlists/p1/tracks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete || r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(playlistTracksResponse{Items: []struct {
			Track trackItem `json:"track"`
		}{{Track: trackItem{ID: "t1", Name: "Track"}}}, Total: 1})
	})
	mux.HandleFunc("/users/me/playlists", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(playlistItem{ID: "p2", Name: "Created"})
	})
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(userProfile{ID: "me"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewClient(Options{TokenProvider: staticTokenProvider{}, BaseURL: srv.URL, Market: "US", Language: "en", Device: "d1"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	if _, err := client.GetTrack(context.Background(), "t1"); err != nil {
		t.Fatalf("track: %v", err)
	}
	if _, err := client.GetAlbum(context.Background(), "a1"); err != nil {
		t.Fatalf("album: %v", err)
	}
	if _, err := client.GetArtist(context.Background(), "ar1"); err != nil {
		t.Fatalf("artist: %v", err)
	}
	if _, err := client.ArtistTopTracks(context.Background(), "ar1", 1); err != nil {
		t.Fatalf("artist top tracks: %v", err)
	}
	if _, err := client.GetPlaylist(context.Background(), "p1"); err != nil {
		t.Fatalf("playlist: %v", err)
	}
	if _, err := client.GetShow(context.Background(), "s1"); err != nil {
		t.Fatalf("show: %v", err)
	}
	if _, err := client.GetEpisode(context.Background(), "e1"); err != nil {
		t.Fatalf("episode: %v", err)
	}
	if _, err := client.Devices(context.Background()); err != nil {
		t.Fatalf("devices: %v", err)
	}
	if err := client.Transfer(context.Background(), "d1"); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if err := client.Play(context.Background(), "spotify:track:t1"); err != nil {
		t.Fatalf("play: %v", err)
	}
	if err := client.Pause(context.Background()); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if err := client.Next(context.Background()); err != nil {
		t.Fatalf("next: %v", err)
	}
	if err := client.Previous(context.Background()); err != nil {
		t.Fatalf("prev: %v", err)
	}
	if err := client.QueueAdd(context.Background(), "spotify:track:t1"); err != nil {
		t.Fatalf("queue add: %v", err)
	}
	if _, err := client.Queue(context.Background()); err != nil {
		t.Fatalf("queue: %v", err)
	}
	if _, _, err := client.LibraryTracks(context.Background(), 1, 0); err != nil {
		t.Fatalf("library tracks: %v", err)
	}
	if _, _, err := client.LibraryAlbums(context.Background(), 1, 0); err != nil {
		t.Fatalf("library albums: %v", err)
	}
	if err := client.LibraryModify(context.Background(), "/me/tracks", []string{"t1"}, http.MethodPut); err != nil {
		t.Fatalf("library modify: %v", err)
	}
	if err := client.FollowArtists(context.Background(), []string{"ar1"}, http.MethodPut); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if _, _, _, err := client.FollowedArtists(context.Background(), 1, ""); err != nil {
		t.Fatalf("followed: %v", err)
	}
	if _, _, err := client.Playlists(context.Background(), 1, 0); err != nil {
		t.Fatalf("playlists: %v", err)
	}
	if _, _, err := client.PlaylistTracks(context.Background(), "p1", 1, 0); err != nil {
		t.Fatalf("playlist tracks: %v", err)
	}
	if _, err := client.CreatePlaylist(context.Background(), "Created", true, false); err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	if err := client.RemoveTracks(context.Background(), "p1", []string{"spotify:track:t1"}); err != nil {
		t.Fatalf("remove tracks: %v", err)
	}
}
