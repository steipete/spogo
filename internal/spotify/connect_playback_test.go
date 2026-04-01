package spotify

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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
	if err := client.Play(context.Background(), "spotify:playlist:abc"); err != nil {
		t.Fatalf("play playlist: %v", err)
	}
	if err := client.Play(context.Background(), "spotify:album:xyz"); err != nil {
		t.Fatalf("play album: %v", err)
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

func TestConnectTransferFallsBackToWebAPIWithoutOriginDevice(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{
				"name":        "Desk",
				"device_type": "computer",
			},
		},
		"player_state": map[string]any{
			"is_paused": true,
		},
	}
	var sawWebTransfer bool
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodPut && req.URL.Path == "/v1/me/player":
			sawWebTransfer = true
			return textResponse(http.StatusNoContent, ""), nil
		case req.Method == http.MethodPost:
			t.Fatalf("unexpected connect command: %s", req.URL.Path)
			return nil, nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	client := newRegisteredConnectClientForTests(transport)
	webClient, err := NewClient(Options{
		TokenProvider: staticTokenProvider{},
		HTTPClient:    client.client,
	})
	if err != nil {
		t.Fatalf("new web client: %v", err)
	}
	client.web = webClient

	if err := client.Transfer(context.Background(), "device-1"); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if !sawWebTransfer {
		t.Fatalf("expected web transfer fallback")
	}
}

func TestConnectPlayFallsBackToWebAPIWithoutActiveDevice(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{
				"name":        "Desk",
				"device_type": "computer",
			},
		},
		"player_state": map[string]any{
			"is_paused": true,
		},
	}
	var sawWebPlay bool
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodPut && req.URL.Path == "/v1/me/player/play":
			sawWebPlay = true
			return textResponse(http.StatusNoContent, ""), nil
		case req.Method == http.MethodPost:
			t.Fatalf("unexpected connect command: %s", req.URL.Path)
			return nil, nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})
	client := newRegisteredConnectClientForTests(transport)
	webClient, err := NewClient(Options{
		TokenProvider: staticTokenProvider{},
		HTTPClient:    client.client,
	})
	if err != nil {
		t.Fatalf("new web client: %v", err)
	}
	client.web = webClient

	if err := client.Play(context.Background(), "spotify:track:abc"); err != nil {
		t.Fatalf("play: %v", err)
	}
	if !sawWebPlay {
		t.Fatalf("expected web play fallback")
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

func TestConnectPlayContextURIPayload(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{"name": "Desk", "device_type": "computer"},
		},
		"player_state": map[string]any{
			"is_paused":   false,
			"position_ms": 0,
		},
		"active_device_id": "device-1",
	}
	var capturedBody string
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/devices/hobs_"):
			return jsonResponse(http.StatusOK, statePayload), nil
		case req.Method == http.MethodPost:
			b, _ := io.ReadAll(req.Body)
			capturedBody = string(b)
			return textResponse(http.StatusOK, "ok"), nil
		default:
			return textResponse(http.StatusNotFound, "missing"), nil
		}
	})

	newClient := func() *ConnectClient {
		c := newConnectClientForTests(transport)
		c.session.connectDeviceID = "device"
		c.session.connectionID = "conn"
		c.session.registeredAt = time.Now()
		return c
	}

	// Context URI (playlist) — must use "context" field, not "track_uri"
	capturedBody = ""
	if err := newClient().Play(context.Background(), "spotify:playlist:pl1"); err != nil {
		t.Fatalf("play playlist: %v", err)
	}
	if !strings.Contains(capturedBody, `"context"`) {
		t.Errorf("playlist play: expected context field in body, got: %s", capturedBody)
	}
	if strings.Contains(capturedBody, `"track_uri"`) {
		t.Errorf("playlist play: unexpected track_uri in body: %s", capturedBody)
	}

	// Track URI — must use "track_uri" and also include "context" (track as its own context)
	capturedBody = ""
	if err := newClient().Play(context.Background(), "spotify:track:t1"); err != nil {
		t.Fatalf("play track: %v", err)
	}
	if !strings.Contains(capturedBody, `"track_uri"`) {
		t.Errorf("track play: expected track_uri field in body, got: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, `"context"`) {
		t.Errorf("track play: expected context field in body, got: %s", capturedBody)
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
