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
