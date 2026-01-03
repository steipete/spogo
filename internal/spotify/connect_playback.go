package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/coder/websocket"
)

const (
	connectStateBase   = "https://gue1-spclient.spotify.com/connect-state/v1"
	trackPlaybackBase  = "https://gue1-spclient.spotify.com/track-playback/v1"
	connectionTTL      = 10 * time.Minute
	connectDeviceName  = "spogo"
	connectDeviceModel = "web_player"
)

var dealerURL = "wss://dealer.spotify.com/"

func (c *ConnectClient) playback(ctx context.Context) (PlaybackStatus, error) {
	state, err := c.connectState(ctx)
	if err != nil {
		return PlaybackStatus{}, err
	}
	return mapPlaybackStatus(state), nil
}

func (c *ConnectClient) devices(ctx context.Context) ([]Device, error) {
	state, err := c.connectState(ctx)
	if err != nil {
		return nil, err
	}
	return mapDevices(state), nil
}

func (c *ConnectClient) transfer(ctx context.Context, deviceID string) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	fromID := state.originDeviceID
	if fromID == "" {
		fromID = state.activeDeviceID
	}
	if fromID == "" {
		return errors.New("missing origin device id")
	}
	return c.sendConnectCommand(ctx, fmt.Sprintf("%s/connect/transfer/from/%s/to/%s", connectStateBase, fromID, deviceID), map[string]any{
		"transfer_options": map[string]any{
			"restore_paused": "resume",
		},
		"command_id": randomHex(32),
	})
}

func (c *ConnectClient) play(ctx context.Context, uri string) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	if uri == "" {
		return c.sendPlayerCommand(ctx, state, "resume", nil)
	}
	payload := map[string]any{
		"command": map[string]any{
			"endpoint": "play",
			"logging_params": map[string]any{
				"command_id": randomHex(32),
			},
			"options": map[string]any{
				"skip_to": map[string]any{
					"track_uri": uri,
				},
			},
		},
	}
	return c.sendPlayerCommand(ctx, state, "play", payload)
}

func (c *ConnectClient) pause(ctx context.Context) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	return c.sendPlayerCommand(ctx, state, "pause", nil)
}

func (c *ConnectClient) next(ctx context.Context) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	return c.sendPlayerCommand(ctx, state, "skip_next", nil)
}

func (c *ConnectClient) previous(ctx context.Context) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	return c.sendPlayerCommand(ctx, state, "skip_prev", nil)
}

func (c *ConnectClient) seek(ctx context.Context, positionMS int) error {
	if positionMS < 0 {
		positionMS = 0
	}
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"command": map[string]any{
			"endpoint": "seek_to",
			"value":    positionMS,
			"logging_params": map[string]any{
				"command_id": randomHex(32),
			},
		},
	}
	return c.sendPlayerCommand(ctx, state, "seek_to", payload)
}

func (c *ConnectClient) volume(ctx context.Context, volume int) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	fromID := state.originDeviceID
	if fromID == "" {
		fromID = state.activeDeviceID
	}
	if fromID == "" || state.activeDeviceID == "" {
		return errors.New("missing device id")
	}
	value := int(float64(volume) / 100 * 65535)
	return c.sendConnectCommand(ctx, fmt.Sprintf("%s/connect/volume/from/%s/to/%s", connectStateBase, fromID, state.activeDeviceID), map[string]any{
		"volume": value,
	})
}

func (c *ConnectClient) shuffle(ctx context.Context, enabled bool) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"command": map[string]any{
			"endpoint": "set_shuffling_context",
			"value":    enabled,
			"logging_params": map[string]any{
				"command_id": randomHex(32),
			},
		},
	}
	return c.sendPlayerCommand(ctx, state, "set_shuffling_context", payload)
}

func (c *ConnectClient) repeat(ctx context.Context, mode string) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	mode = strings.ToLower(mode)
	payload := map[string]any{
		"command": map[string]any{
			"endpoint": "set_options",
			"logging_params": map[string]any{
				"command_id": randomHex(32),
			},
		},
	}
	switch mode {
	case "track":
		payload["command"].(map[string]any)["repeating_track"] = true
		payload["command"].(map[string]any)["repeating_context"] = false
	case "context":
		payload["command"].(map[string]any)["repeating_track"] = false
		payload["command"].(map[string]any)["repeating_context"] = true
	default:
		payload["command"].(map[string]any)["repeating_track"] = false
		payload["command"].(map[string]any)["repeating_context"] = false
	}
	return c.sendPlayerCommand(ctx, state, "set_options", payload)
}

func (c *ConnectClient) queueAdd(ctx context.Context, uri string) error {
	state, err := c.connectState(ctx)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"command": map[string]any{
			"endpoint": "add_to_queue",
			"track": map[string]any{
				"uri": uri,
			},
			"logging_params": map[string]any{
				"command_id": randomHex(32),
			},
		},
	}
	return c.sendPlayerCommand(ctx, state, "add_to_queue", payload)
}

func (c *ConnectClient) queue(ctx context.Context) (Queue, error) {
	state, err := c.connectState(ctx)
	if err != nil {
		return Queue{}, err
	}
	return mapQueue(state), nil
}

type connectState struct {
	raw            map[string]any
	playerState    map[string]any
	devices        map[string]any
	activeDeviceID string
	originDeviceID string
}

func (c *ConnectClient) connectState(ctx context.Context) (connectState, error) {
	auth, err := c.session.auth(ctx)
	if err != nil {
		return connectState{}, err
	}
	if err := c.ensureConnectDevice(ctx, auth); err != nil {
		return connectState{}, err
	}
	c.session.mu.Lock()
	deviceID := c.session.connectDeviceID
	connectionID := c.session.connectionID
	c.session.mu.Unlock()
	payload := map[string]any{
		"member_type": "CONNECT_STATE",
		"device": map[string]any{
			"device_info": map[string]any{
				"capabilities": map[string]any{
					"can_be_player":           false,
					"hidden":                  true,
					"needs_full_player_state": true,
				},
			},
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/devices/hobs_%s", connectStateBase, deviceID), encodeJSON(payload))
	if err != nil {
		return connectState{}, err
	}
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Client-Token", auth.ClientToken)
	req.Header.Set("Spotify-App-Version", connectVersion(auth))
	req.Header.Set("User-Agent", defaultUserAgent())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("app-platform", "WebPlayer")
	if connectionID != "" {
		req.Header.Set("x-spotify-connection-id", connectionID)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return connectState{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return connectState{}, apiErrorFromResponse(resp)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return connectState{}, err
	}
	state := connectState{raw: raw}
	if devices, ok := raw["devices"].(map[string]any); ok {
		state.devices = devices
	}
	if player, ok := raw["player_state"].(map[string]any); ok {
		state.playerState = player
	}
	if active, ok := raw["active_device_id"].(string); ok {
		state.activeDeviceID = active
	}
	if origin := mapPlayOriginID(state.playerState); origin != "" {
		state.originDeviceID = origin
	}
	return state, nil
}

func (c *ConnectClient) ensureConnectDevice(ctx context.Context, auth connectAuth) error {
	c.session.mu.Lock()
	if c.session.connectDeviceID == "" {
		c.session.connectDeviceID = randomHex(32)
	}
	needs := c.session.connectionID == "" || time.Since(c.session.registeredAt) > connectionTTL
	c.session.mu.Unlock()
	if !needs {
		return nil
	}
	connectionID, err := getConnectionID(ctx, auth.AccessToken)
	if err != nil {
		return err
	}
	if err := c.registerDevice(ctx, auth, connectionID); err != nil {
		return err
	}
	c.session.mu.Lock()
	c.session.connectionID = connectionID
	c.session.registeredAt = time.Now()
	c.session.mu.Unlock()
	return nil
}

func (c *ConnectClient) registerDevice(ctx context.Context, auth connectAuth, connectionID string) error {
	c.session.mu.Lock()
	deviceID := c.session.connectDeviceID
	c.session.mu.Unlock()
	payload := map[string]any{
		"device": map[string]any{
			"device_id":           deviceID,
			"device_type":         "computer",
			"brand":               "spotify",
			"model":               connectDeviceModel,
			"name":                connectDeviceName,
			"is_group":            false,
			"metadata":            map[string]any{},
			"platform_identifier": fmt.Sprintf("web_player %s;spogo", runtime.GOOS),
			"capabilities": map[string]any{
				"change_volume":            true,
				"supports_file_media_type": true,
				"enable_play_token":        true,
				"play_token_lost_behavior": "pause",
				"disable_connect":          false,
				"audio_podcasts":           true,
				"video_playback":           true,
				"manifest_formats": []string{
					"file_ids_mp3",
					"file_urls_mp3",
					"file_ids_mp4",
					"manifest_ids_video",
				},
			},
		},
		"outro_endcontent_snooping": false,
		"connection_id":             connectionID,
		"client_version":            connectVersion(auth),
		"volume":                    65535,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, trackPlaybackBase+"/devices", encodeJSON(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Client-Token", auth.ClientToken)
	req.Header.Set("Spotify-App-Version", connectVersion(auth))
	req.Header.Set("User-Agent", defaultUserAgent())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("app-platform", "WebPlayer")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiErrorFromResponse(resp)
	}
	return nil
}

func (c *ConnectClient) sendPlayerCommand(ctx context.Context, state connectState, endpoint string, payload map[string]any) error {
	if payload == nil {
		payload = map[string]any{
			"command": map[string]any{
				"endpoint": endpoint,
				"logging_params": map[string]any{
					"command_id": randomHex(32),
				},
			},
		}
	}
	fromID := state.originDeviceID
	if fromID == "" {
		fromID = state.activeDeviceID
	}
	if fromID == "" || state.activeDeviceID == "" {
		return errors.New("missing device id")
	}
	url := fmt.Sprintf("%s/player/command/from/%s/to/%s", connectStateBase, fromID, state.activeDeviceID)
	return c.sendConnectCommand(ctx, url, payload)
}

func (c *ConnectClient) sendConnectCommand(ctx context.Context, url string, payload map[string]any) error {
	auth, err := c.session.auth(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, encodeJSON(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Client-Token", auth.ClientToken)
	req.Header.Set("Spotify-App-Version", connectVersion(auth))
	req.Header.Set("User-Agent", defaultUserAgent())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("app-platform", "WebPlayer")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiErrorFromResponse(resp)
	}
	return nil
}

func getConnectionID(ctx context.Context, accessToken string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	url := dealerURL
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
		if strings.HasSuffix(url, "?") || strings.HasSuffix(url, "&") {
			sep = ""
		}
	}
	url += sep + "access_token=" + accessToken
	conn, resp, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"User-Agent": []string{defaultUserAgent()},
		},
	})
	if err != nil {
		return "", err
	}
	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()
	_, data, err := conn.Read(ctx)
	if err != nil {
		return "", err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	headers, ok := payload["headers"].(map[string]any)
	if !ok {
		return "", errors.New("missing headers")
	}
	for key, value := range headers {
		if !strings.EqualFold(key, "Spotify-Connection-Id") {
			continue
		}
		if id, ok := value.(string); ok && id != "" {
			return id, nil
		}
	}
	return "", errors.New("missing connection id")
}
