package spotify

import "testing"

func TestMapPlaybackStatusAndDevices(t *testing.T) {
	state := connectState{
		activeDeviceID: "device-1",
		devices: map[string]any{
			"device-1": map[string]any{
				"name":           "Desk",
				"device_type":    "computer",
				"volume_percent": 40,
			},
			"device-2": map[string]any{
				"device_name": "Phone",
				"device_type": "smartphone",
				"volume":      80,
			},
		},
		playerState: map[string]any{
			"is_paused":   true,
			"position_ms": 1200,
			"shuffle":     true,
			"repeat":      "context",
			"track": map[string]any{
				"uri":  "spotify:track:abc",
				"name": "Song",
			},
		},
	}
	status := mapPlaybackStatus(state)
	if status.IsPlaying {
		t.Fatalf("expected paused")
	}
	if status.Device.ID != "device-1" || status.Device.Name != "Desk" {
		t.Fatalf("unexpected device: %#v", status.Device)
	}
	if status.Item == nil || status.Item.URI != "spotify:track:abc" {
		t.Fatalf("expected item")
	}
}

func TestMapDevicesUsesLabelFallback(t *testing.T) {
	state := connectState{
		activeDeviceID: "d1",
		devices: map[string]any{
			"d1": map[string]any{
				"label":       "Sony TV",
				"device_type": "tv",
				"is_active":   true,
			},
		},
		playerState: map[string]any{"is_paused": true},
	}
	status := mapPlaybackStatus(state)
	if status.Device.ID != "d1" || status.Device.Name != "Sony TV" || !status.Device.Active {
		t.Fatalf("unexpected device: %#v", status.Device)
	}
}

func TestMapQueue(t *testing.T) {
	state := connectState{
		playerState: map[string]any{
			"track": map[string]any{
				"uri":  "spotify:track:now",
				"name": "Now",
			},
			"next_tracks": []any{
				map[string]any{"track": map[string]any{"uri": "spotify:track:next", "name": "Next"}},
			},
		},
	}
	queue := mapQueue(state)
	if queue.CurrentlyPlaying == nil || queue.CurrentlyPlaying.URI != "spotify:track:now" {
		t.Fatalf("expected current item")
	}
	if len(queue.Queue) != 1 || queue.Queue[0].URI != "spotify:track:next" {
		t.Fatalf("expected next item")
	}
}

func TestExtractPlaybackTrackContext(t *testing.T) {
	player := map[string]any{
		"context_uri": "spotify:album:abc",
	}
	item := extractPlaybackTrack(player)
	if item.URI != "spotify:album:abc" || item.Type != "album" {
		t.Fatalf("unexpected item: %#v", item)
	}
}

func TestExtractPlaybackTrackCurrent(t *testing.T) {
	player := map[string]any{
		"current_track": map[string]any{
			"uri":  "spotify:track:xyz",
			"name": "Song",
		},
	}
	item := extractPlaybackTrack(player)
	if item.URI != "spotify:track:xyz" {
		t.Fatalf("unexpected item: %#v", item)
	}
}

func TestExtractPlaybackTrackEnrichesFromTrackMetadata(t *testing.T) {
	player := map[string]any{
		"track": map[string]any{
			"uri": "spotify:track:xyz",
		},
		"track_metadata": map[string]any{
			"title":       "Song",
			"artist_name": "Artist",
			"album_title": "Album",
		},
	}
	item := extractPlaybackTrack(player)
	if item.URI != "spotify:track:xyz" || item.Name != "Song" || item.Album != "Album" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if len(item.Artists) != 1 || item.Artists[0] != "Artist" {
		t.Fatalf("unexpected artists: %#v", item.Artists)
	}
}

func TestExtractPlaybackTrackTrackWindowShape(t *testing.T) {
	player := map[string]any{
		"track_window": map[string]any{
			"current_track": map[string]any{
				"uri":  "spotify:track:win",
				"name": "Window Song",
			},
		},
	}
	item := extractPlaybackTrack(player)
	if item.URI != "spotify:track:win" || item.Name != "Window Song" {
		t.Fatalf("unexpected item: %#v", item)
	}
}
