package spotify

import (
	"encoding/json"
	"testing"
)

func TestMapSearchItemTrack(t *testing.T) {
	raw := json.RawMessage(`{"id":"t1","uri":"spotify:track:t1","name":"Song","duration_ms":1000,"explicit":false,"is_playable":true,"album":{"name":"Album"},"artists":[{"name":"Artist"}]}`)
	item, err := mapSearchItem("track", raw)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if item.Name != "Song" || item.Type != "track" {
		t.Fatalf("unexpected item: %#v", item)
	}
}

func TestMapArtist(t *testing.T) {
	item := mapArtist(artistItem{ID: "a1", Name: "Artist", Followers: struct {
		Total int `json:"total"`
	}{Total: 10}})
	if item.Followers != 10 {
		t.Fatalf("followers: %d", item.Followers)
	}
}

func TestMapSearchItemUnsupported(t *testing.T) {
	if _, err := mapSearchItem("unknown", json.RawMessage(`{}`)); err == nil {
		t.Fatalf("expected error")
	}
}

func TestMapSearchItemAlbum(t *testing.T) {
	raw := json.RawMessage(`{"id":"a1","uri":"spotify:album:a1","name":"Album","artists":[{"name":"Artist"}],"album_type":"album"}`)
	if _, err := mapSearchItem("album", raw); err != nil {
		t.Fatalf("album: %v", err)
	}
}

func TestMapSearchItemArtist(t *testing.T) {
	raw := json.RawMessage(`{"id":"ar1","uri":"spotify:artist:ar1","name":"Artist","followers":{"total":1}}`)
	if _, err := mapSearchItem("artist", raw); err != nil {
		t.Fatalf("artist: %v", err)
	}
}

func TestMapSearchItemPlaylist(t *testing.T) {
	raw := json.RawMessage(`{"id":"p1","uri":"spotify:playlist:p1","name":"Playlist","tracks":{"total":1},"owner":{"display_name":"Owner"}}`)
	if _, err := mapSearchItem("playlist", raw); err != nil {
		t.Fatalf("playlist: %v", err)
	}
}

func TestMapSearchItemShow(t *testing.T) {
	raw := json.RawMessage(`{"id":"s1","uri":"spotify:show:s1","name":"Show","publisher":"Pub","total_episodes":1}`)
	if _, err := mapSearchItem("show", raw); err != nil {
		t.Fatalf("show: %v", err)
	}
}

func TestMapSearchItemEpisode(t *testing.T) {
	raw := json.RawMessage(`{"id":"e1","uri":"spotify:episode:e1","name":"Episode","duration_ms":1000}`)
	if _, err := mapSearchItem("episode", raw); err != nil {
		t.Fatalf("episode: %v", err)
	}
}

func TestExternalURL(t *testing.T) {
	if externalURL(nil) != "" {
		t.Fatalf("expected empty")
	}
	url := externalURL(map[string]string{"foo": "bar"})
	if url != "bar" {
		t.Fatalf("expected fallback url")
	}
}
