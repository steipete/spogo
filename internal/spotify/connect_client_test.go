package spotify

import (
	"context"
	"net/http"
	"testing"
)

func TestNewConnectClient(t *testing.T) {
	if _, err := NewConnectClient(ConnectOptions{}); err == nil {
		t.Fatalf("expected error")
	}
	_, err := NewConnectClient(ConnectOptions{Source: cookieSourceStub{cookies: []*http.Cookie{{Name: "sp_dc", Value: "token"}}}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
}

func TestConnectInfoOperations(t *testing.T) {
	payloads := map[string]map[string]any{
		"getTrack": {
			"data": map[string]any{"track": map[string]any{"uri": "spotify:track:t1", "name": "Song"}},
		},
		"getAlbum": {
			"data": map[string]any{"album": map[string]any{"uri": "spotify:album:a1", "name": "Album"}},
		},
		"queryArtistOverview": {
			"data": map[string]any{"artist": map[string]any{"uri": "spotify:artist:ar1", "name": "Artist"}},
		},
		"fetchPlaylist": {
			"data": map[string]any{"playlist": map[string]any{"uri": "spotify:playlist:p1", "name": "Playlist"}},
		},
		"queryPodcastEpisodes": {
			"data": map[string]any{"show": map[string]any{"uri": "spotify:show:s1", "name": "Show"}},
		},
		"getEpisodeOrChapter": {
			"data": map[string]any{"episode": map[string]any{"uri": "spotify:episode:e1", "name": "Episode"}},
		},
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		op := req.URL.Query().Get("operationName")
		payload, ok := payloads[op]
		if !ok {
			return textResponse(http.StatusNotFound, "missing"), nil
		}
		return jsonResponse(http.StatusOK, payload), nil
	})
	client := newConnectClientForTests(transport)
	for op := range payloads {
		client.hashes.hashes[op] = "hash"
	}
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

func TestConnectUnsupported(t *testing.T) {
	client := &ConnectClient{}
	if _, _, err := client.LibraryTracks(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error")
	}
	if _, _, err := client.LibraryAlbums(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.LibraryModify(context.Background(), "", nil, ""); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.FollowArtists(context.Background(), nil, ""); err == nil {
		t.Fatalf("expected error")
	}
	if _, _, _, err := client.FollowedArtists(context.Background(), 1, ""); err == nil {
		t.Fatalf("expected error")
	}
	if _, _, err := client.Playlists(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error")
	}
	if _, _, err := client.PlaylistTracks(context.Background(), "p1", 1, 0); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.CreatePlaylist(context.Background(), "name", false, false); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.AddTracks(context.Background(), "p1", nil); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.RemoveTracks(context.Background(), "p1", nil); err == nil {
		t.Fatalf("expected error")
	}
}
