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
		device := Device{
			ID:     id,
			Name:   name,
			Type:   getString(deviceMap, "device_type"),
			Active: id == state.activeDeviceID,
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
			if item, ok := extractItem(entry, "track"); ok {
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
	for _, key := range []string{"track", "item", "current_track"} {
		if raw, ok := player[key]; ok {
			if item, ok := extractItem(raw, "track"); ok {
				return item
			}
		}
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
