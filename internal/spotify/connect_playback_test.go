package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestConnectPlaybackCommands(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{
				"name":        "Desk",
				"device_type": "computer",
				"volume":      10,
			},
		},
		"player_state": map[string]any{
			"is_paused":   false,
			"position_ms": 100,
			"track": map[string]any{
				"uri":  "spotify:track:abc",
				"name": "Song",
			},
		},
		"active_device_id": "device-1",
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodPost:
			return textResponse(http.StatusOK, "ok"), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	if _, err := client.Playback(context.Background()); err != nil {
		t.Fatalf("playback: %v", err)
	}
	if _, err := client.Devices(context.Background()); err != nil {
		t.Fatalf("devices: %v", err)
	}
	if err := client.Play(context.Background(), ""); err != nil {
		t.Fatalf("play resume: %v", err)
	}
	if err := client.Play(context.Background(), "spotify:track:abc"); err != nil {
		t.Fatalf("play uri: %v", err)
	}
	if err := client.Pause(context.Background()); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if err := client.Next(context.Background()); err != nil {
		t.Fatalf("next: %v", err)
	}
	if err := client.Previous(context.Background()); err != nil {
		t.Fatalf("previous: %v", err)
	}
	if err := client.Seek(context.Background(), -1); err != nil {
		t.Fatalf("seek: %v", err)
	}
	if err := client.Volume(context.Background(), -5); err != nil {
		t.Fatalf("volume: %v", err)
	}
	if err := client.Volume(context.Background(), 200); err != nil {
		t.Fatalf("volume high: %v", err)
	}
	if err := client.Shuffle(context.Background(), true); err != nil {
		t.Fatalf("shuffle: %v", err)
	}
	if err := client.Repeat(context.Background(), "track"); err != nil {
		t.Fatalf("repeat track: %v", err)
	}
	if err := client.Repeat(context.Background(), "context"); err != nil {
		t.Fatalf("repeat context: %v", err)
	}
	if err := client.Repeat(context.Background(), "off"); err != nil {
		t.Fatalf("repeat off: %v", err)
	}
	if err := client.QueueAdd(context.Background(), "spotify:track:abc"); err != nil {
		t.Fatalf("queue add: %v", err)
	}
	if _, err := client.Queue(context.Background()); err != nil {
		t.Fatalf("queue: %v", err)
	}
	if err := client.Transfer(context.Background(), "device-1"); err != nil {
		t.Fatalf("transfer: %v", err)
	}
}

func TestConnectPlaybackHonorsTargetDeviceSelector(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{
				"name":        "Desk",
				"device_type": "computer",
			},
			"device-2": map[string]any{
				"name":        "Phone",
				"device_type": "smartphone",
			},
		},
		"player_state":      map[string]any{"is_paused": false},
		"active_device_id":  "device-1",
		"connection_id":     "conn",
		"last_command_sent": "",
	}
	var transferCalls, commandCalls int
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/connect/transfer/from/"):
			transferCalls++
			return textResponse(http.StatusOK, "ok"), nil
		case req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/player/command/from/"):
			commandCalls++
			if !strings.Contains(req.URL.Path, "/to/device-2") {
				return textResponse(http.StatusBadRequest, "wrong target"), nil
			}
			return textResponse(http.StatusOK, "ok"), nil
		default:
			return textResponse(http.StatusOK, "ok"), nil
		}
	})
	client := newConnectClientForTests(transport)
	client.device = "Phone"
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	if err := client.Pause(context.Background()); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if transferCalls != 1 {
		t.Fatalf("expected transfer call, got %d", transferCalls)
	}
	if commandCalls != 1 {
		t.Fatalf("expected player command call, got %d", commandCalls)
	}
}

type tokenProviderStub struct{}

func (tokenProviderStub) Token(context.Context) (Token, error) {
	return Token{AccessToken: "token", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func TestConnectPlaybackHydratesDeviceNameViaWebDevices(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"sony-1": map[string]any{},
		},
		"player_state": map[string]any{
			"is_paused": true,
			"track": map[string]any{
				"uri": "spotify:track:abc",
			},
		},
		"active_device_id": "sony-1",
	}
	var webPlaybackCalls, webDevicesCalls int
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodGet && req.URL.Host == "api.spotify.com" && req.URL.Path == "/v1/me/player":
			webPlaybackCalls++
			// Simulate Web API playback returning a device id but missing name (observed for some sessions).
			return jsonResponse(http.StatusOK, map[string]any{
				"is_playing":             false,
				"progress_ms":            0,
				"shuffle_state":          false,
				"repeat_state":           "off",
				"device":                 map[string]any{"id": "sony-1", "name": "", "type": "tv", "volume_percent": 0, "is_active": true, "is_restricted": false},
				"item":                   map[string]any{"id": "abc", "name": "Song", "uri": "spotify:track:abc", "album": map[string]any{"name": "Album"}, "artists": []any{map[string]any{"name": "Artist"}}},
				"currently_playing_type": "track",
			}), nil
		case req.Method == http.MethodGet && req.URL.Host == "api.spotify.com" && req.URL.Path == "/v1/me/player/devices":
			webDevicesCalls++
			return jsonResponse(http.StatusOK, deviceResponse{
				Devices: []deviceItem{
					{ID: "sony-1", Name: "Sony TV", Type: "tv", Volume: 0, Active: true, Restricted: false},
				},
			}), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	connectClient := newConnectClientForTests(transport)
	connectClient.session.connectDeviceID = "device"
	connectClient.session.connectionID = "conn"
	connectClient.session.registeredAt = time.Now()

	webClient, err := NewClient(Options{
		TokenProvider: tokenProviderStub{},
		HTTPClient:    &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("web client: %v", err)
	}
	connectClient.web = webClient

	status, err := connectClient.Playback(context.Background())
	if err != nil {
		t.Fatalf("playback: %v", err)
	}
	if webPlaybackCalls == 0 || webDevicesCalls == 0 {
		t.Fatalf("expected web hydration calls, playback=%d devices=%d", webPlaybackCalls, webDevicesCalls)
	}
	if status.Device.ID != "sony-1" || status.Device.Name != "Sony TV" {
		t.Fatalf("unexpected device: %#v", status.Device)
	}
}

func TestConnectPlaybackLastResortDeviceFromSingleActiveWebDevice(t *testing.T) {
	statePayload := map[string]any{
		// Device payload is present but barren; no active_device_id and no active flags.
		"devices": map[string]any{
			"sony-1": map[string]any{},
		},
		"player_state": map[string]any{
			"is_paused": true,
			"track": map[string]any{
				"uri": "spotify:track:abc",
			},
		},
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodGet && req.URL.Host == "api.spotify.com" && req.URL.Path == "/v1/me/player":
			// Don't help with device metadata.
			return jsonResponse(http.StatusOK, map[string]any{
				"is_playing":    false,
				"progress_ms":   0,
				"shuffle_state": false,
				"repeat_state":  "off",
				"device":        map[string]any{"id": "", "name": "", "type": "", "volume_percent": 0, "is_active": false, "is_restricted": false},
				"item":          map[string]any{"id": "abc", "name": "Song", "uri": "spotify:track:abc"},
			}), nil
		case req.Method == http.MethodGet && req.URL.Host == "api.spotify.com" && req.URL.Path == "/v1/me/player/devices":
			return jsonResponse(http.StatusOK, deviceResponse{
				Devices: []deviceItem{
					{ID: "web-sony", Name: "Sony TV", Type: "tv", Volume: 0, Active: true, Restricted: false},
				},
			}), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	connectClient := newConnectClientForTests(transport)
	connectClient.session.connectDeviceID = "device"
	connectClient.session.connectionID = "conn"
	connectClient.session.registeredAt = time.Now()
	connectClient.hashes = newHashResolver(&http.Client{Transport: transport}, connectClient.session)

	webClient, err := NewClient(Options{
		TokenProvider: tokenProviderStub{},
		HTTPClient:    &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("web client: %v", err)
	}
	connectClient.web = webClient

	status, err := connectClient.Playback(context.Background())
	if err != nil {
		t.Fatalf("playback: %v", err)
	}
	if status.Device.Name != "Sony TV" || status.Device.ID != "web-sony" {
		t.Fatalf("unexpected device: %#v", status.Device)
	}
}

func TestConnectPlaybackActiveDeviceFromDevices(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{
				"name":        "Desk",
				"device_type": "computer",
				"is_active":   true,
			},
		},
		"player_state": map[string]any{
			"is_paused": true,
		},
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodPost:
			return textResponse(http.StatusOK, "ok"), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	devices, err := client.Devices(context.Background())
	if err != nil {
		t.Fatalf("devices: %v", err)
	}
	if len(devices) != 1 || !devices[0].Active {
		t.Fatalf("expected active device: %#v", devices)
	}
	if err := client.Transfer(context.Background(), "device-1"); err != nil {
		t.Fatalf("transfer: %v", err)
	}
}

func TestSendPlayerCommandMissingDevice(t *testing.T) {
	client := newConnectClientForTests(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return textResponse(http.StatusOK, ""), nil
	}))
	err := client.sendPlayerCommand(context.Background(), connectState{}, "pause", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRandomHexAndOrigin(t *testing.T) {
	if randomHex(0) != "" {
		t.Fatalf("expected empty")
	}
	value := randomHex(6)
	if len(value) != 6 {
		t.Fatalf("unexpected length: %s", value)
	}
	player := map[string]any{"play_origin": map[string]any{"device_identifier": "abc"}}
	if mapPlayOriginID(player) != "abc" {
		t.Fatalf("expected origin id")
	}
	if connectVersion(connectAuth{ClientVersion: "a"}) != "a" {
		t.Fatalf("expected client version fallback")
	}
	if connectVersion(connectAuth{ClientVersion: "a", ConnectVersion: "b"}) != "b" {
		t.Fatalf("expected connect version")
	}
}

func TestGetConnectionID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("accept: %v", err)
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		payload := map[string]any{
			"headers": map[string]any{
				"Spotify-Connection-Id": "conn-id",
			},
		}
		data, _ := json.Marshal(payload)
		if err := conn.Write(r.Context(), websocket.MessageText, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()

	prev := dealerURL
	dealerURL = "ws" + strings.TrimPrefix(srv.URL, "http")
	t.Cleanup(func() { dealerURL = prev })

	id, err := getConnectionID(context.Background(), "token")
	if err != nil {
		t.Fatalf("getConnectionID: %v", err)
	}
	if id != "conn-id" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestEnsureConnectDeviceRegisters(t *testing.T) {
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("accept: %v", err)
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		payload := map[string]any{
			"headers": map[string]any{
				"Spotify-Connection-Id": "conn-xyz",
			},
		}
		data, _ := json.Marshal(payload)
		if err := conn.Write(r.Context(), websocket.MessageText, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer wsServer.Close()

	prev := dealerURL
	dealerURL = "ws" + strings.TrimPrefix(wsServer.URL, "http")
	t.Cleanup(func() { dealerURL = prev })

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPost {
			return textResponse(http.StatusOK, "ok"), nil
		}
		return textResponse(http.StatusOK, "ok"), nil
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = ""
	client.session.registeredAt = time.Time{}

	auth := connectAuth{
		AccessToken:   "access",
		ClientToken:   "client-token",
		ClientVersion: "1.0.0",
	}
	if err := client.ensureConnectDevice(context.Background(), auth); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if client.session.connectionID == "" {
		t.Fatalf("expected connection id")
	}
}

func TestGetConnectionIDMissingHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("accept: %v", err)
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		payload := map[string]any{
			"headers": map[string]any{
				"Other": "nope",
			},
		}
		data, _ := json.Marshal(payload)
		if err := conn.Write(r.Context(), websocket.MessageText, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()

	prev := dealerURL
	dealerURL = "ws" + strings.TrimPrefix(srv.URL, "http")
	t.Cleanup(func() { dealerURL = prev })

	if _, err := getConnectionID(context.Background(), "token"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetConnectionIDBadHeadersType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("accept: %v", err)
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		payload := map[string]any{
			"headers": "bad",
		}
		data, _ := json.Marshal(payload)
		if err := conn.Write(r.Context(), websocket.MessageText, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()

	prev := dealerURL
	dealerURL = "ws" + strings.TrimPrefix(srv.URL, "http")
	t.Cleanup(func() { dealerURL = prev })

	if _, err := getConnectionID(context.Background(), "token"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetConnectionIDDialError(t *testing.T) {
	prev := dealerURL
	dealerURL = "ws://127.0.0.1:1"
	t.Cleanup(func() { dealerURL = prev })

	if _, err := getConnectionID(context.Background(), "token"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetConnectionIDBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("accept: %v", err)
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		if err := conn.Write(r.Context(), websocket.MessageText, []byte("nope")); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()

	prev := dealerURL
	dealerURL = "ws" + strings.TrimPrefix(srv.URL, "http")
	t.Cleanup(func() { dealerURL = prev })

	if _, err := getConnectionID(context.Background(), "token"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestConnectPlaybackErrorPaths(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return textResponse(http.StatusInternalServerError, "fail"), nil
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	if _, err := client.Playback(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.Devices(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.Pause(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.Next(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.Previous(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.Shuffle(context.Background(), false); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.QueueAdd(context.Background(), "spotify:track:abc"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.Queue(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if err := client.Transfer(context.Background(), "device-1"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSendConnectCommandHTTPError(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return textResponse(http.StatusInternalServerError, "fail"), nil
	})
	client := newConnectClientForTests(transport)
	client.session.token = Token{AccessToken: "access", ExpiresAt: time.Now().Add(time.Hour)}
	client.session.clientToken = "ct"
	client.session.clientTokenT = time.Now().Add(time.Hour)
	client.session.clientVer = "1.0.0"
	client.session.deviceID = "device"

	if err := client.sendConnectCommand(context.Background(), "https://example.com", map[string]any{}); err == nil {
		t.Fatalf("expected error")
	}
}
