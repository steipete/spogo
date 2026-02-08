package spotify

import (
	"strings"
)

func mapDevices(state connectState) []Device {
	if state.devices == nil {
		return nil
	}
	devices := make([]Device, 0, len(state.devices))
	for id, raw := range state.devices {
		deviceMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name := getString(deviceMap, "name")
		if name == "" {
			name = getString(deviceMap, "device_name")
		}
		if name == "" {
			name = getString(deviceMap, "deviceName")
		}
		if name == "" {
			name = getString(deviceMap, "label")
		}
		if name == "" {
			name = getString(deviceMap, "friendly_name")
		}
		devType := getString(deviceMap, "device_type")
		if devType == "" {
			devType = getString(deviceMap, "type")
		}
		if devType == "" {
			devType = getString(deviceMap, "deviceType")
		}
		device := Device{
			ID:     id,
			Name:   name,
			Type:   devType,
			Active: id == state.activeDeviceID,
		}
		device.Restricted = getBool(deviceMap, "is_restricted")
		if !device.Active {
			device.Active = getBool(deviceMap, "is_active") || getBool(deviceMap, "is_currently_playing") || getBool(deviceMap, "is_active_device")
		}
		device.Volume = getInt(deviceMap, "volume")
		if device.Volume == 0 {
			device.Volume = getInt(deviceMap, "volume_percent")
		}
		devices = append(devices, device)
	}
	return devices
}

func mapPlaybackStatus(state connectState) PlaybackStatus {
	status := PlaybackStatus{}
	player := state.playerState
	if player == nil {
		return status
	}
	if paused, ok := player["is_paused"].(bool); ok {
		status.IsPlaying = !paused
	} else {
		status.IsPlaying = getBool(player, "is_playing")
	}
	status.ProgressMS = getInt(player, "position_as_of_timestamp")
	if status.ProgressMS == 0 {
		status.ProgressMS = getInt(player, "position_ms")
	}
	status.Shuffle = getBool(player, "shuffle")
	status.Repeat = getString(player, "repeat_mode")
	if status.Repeat == "" {
		status.Repeat = getString(player, "repeat")
	}
	if track := extractPlaybackTrack(player); track.URI != "" {
		status.Item = &track
	}
	// Even if device metadata is missing, preserve the active device id.
	if state.activeDeviceID != "" {
		status.Device.ID = state.activeDeviceID
		status.Device.Active = true
	}
	for _, device := range mapDevices(state) {
		if device.Active {
			status.Device = device
			break
		}
	}
	return status
}

func mapQueue(state connectState) Queue {
	queue := Queue{}
	if state.playerState == nil {
		return queue
	}
	if current := extractPlaybackTrack(state.playerState); current.URI != "" {
		queue.CurrentlyPlaying = &current
	}
	if next, ok := state.playerState["next_tracks"].([]any); ok {
		for _, entry := range next {
			// "next_tracks" can include non-tracks (episodes, ads, etc). Don't hard-filter to track.
			if item, ok := extractItem(entry, ""); ok {
				queue.Queue = append(queue.Queue, item)
			}
		}
	}
	return queue
}

func extractPlaybackTrack(player map[string]any) Item {
	if player == nil {
		return Item{}
	}
	// Common connect-state shapes.
	for _, key := range []string{"track", "item", "current_track"} {
		if raw, ok := player[key]; ok {
			if item, ok := extractItem(raw, ""); ok {
				return enrichPlaybackItem(item, player)
			}
		}
	}
	// Web Playback SDK-style shape.
	if tw, ok := player["track_window"].(map[string]any); ok {
		if raw, ok := tw["current_track"]; ok {
			if item, ok := extractItem(raw, ""); ok {
				return enrichPlaybackItem(item, player)
			}
		}
	}
	// Some payloads split metadata away from the track object.
	if uri := getString(player, "track_uri"); uri != "" && strings.HasPrefix(uri, "spotify:") {
		item := Item{
			URI:  uri,
			ID:   idFromURI(uri),
			Type: typeFromURI(uri),
		}
		return enrichPlaybackItem(item, player)
	}
	for _, key := range []string{"context_uri", "context_uri_string"} {
		if uri, ok := player[key].(string); ok && strings.HasPrefix(uri, "spotify:") {
			item := Item{
				URI:  uri,
				ID:   idFromURI(uri),
				Type: typeFromURI(uri),
			}
			return item
		}
	}
	return Item{}
}

func enrichPlaybackItem(item Item, player map[string]any) Item {
	// If the primary object only contains a URI, try to enrich from adjacent metadata.
	if player == nil {
		return item
	}
	needsName := item.Name == ""
	needsArtists := len(item.Artists) == 0
	needsAlbum := item.Album == ""
	if !needsName && !needsArtists && !needsAlbum {
		return item
	}
	for _, key := range []string{"track_metadata", "metadata"} {
		raw, ok := player[key]
		if !ok {
			continue
		}
		if needsName {
			if name := findFirstName(raw); name != "" {
				item.Name = name
				needsName = false
			}
		}
		if needsArtists {
			if artists := extractArtistNames(raw); len(artists) > 0 {
				item.Artists = artists
				needsArtists = false
			}
		}
		if needsAlbum {
			if album := extractAlbumName(raw); album != "" {
				item.Album = album
				needsAlbum = false
			}
		}
		if !needsName && !needsArtists && !needsAlbum {
			break
		}
	}
	return item
}
