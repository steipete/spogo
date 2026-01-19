package spotify

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestVerifyPlaybackSucceeds(t *testing.T) {
	statePayloads := []map[string]any{
		{
			"devices": map[string]any{
				"device-1": map[string]any{
					"name":        "Desk",
					"device_type": "computer",
				},
			},
			"player_state": map[string]any{
				"is_paused":   false,
				"position_ms": 0,
				"track": map[string]any{
					"uri":  "spotify:track:abc",
					"name": "Song",
				},
			},
			"active_device_id": "device-1",
		},
		{
			"devices": map[string]any{
				"device-1": map[string]any{
					"name":        "Desk",
					"device_type": "computer",
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
		},
	}
	idx := 0
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/connect-state/v1/devices/hobs_") {
			payload := statePayloads[idx]
			if idx < len(statePayloads)-1 {
				idx++
			}
			return jsonResponse(http.StatusOK, payload), nil
		}
		return textResponse(http.StatusNotFound, "missing"), nil
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	status, err := VerifyPlayback(context.Background(), client, 50*time.Millisecond, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if status.ProgressMS <= 0 {
		t.Fatalf("expected progress_ms > 0, got %d", status.ProgressMS)
	}
	if status.Item == nil || status.Item.URI == "" {
		t.Fatalf("expected item, got %#v", status.Item)
	}
	if status.Device.ID == "" {
		t.Fatalf("expected device")
	}
}

func TestVerifyPlaybackFailsOnStuckProgress(t *testing.T) {
	statePayload := map[string]any{
		"devices": map[string]any{
			"device-1": map[string]any{
				"name":        "Desk",
				"device_type": "computer",
			},
		},
		"player_state": map[string]any{
			"is_paused":   false,
			"position_ms": 0,
			"track": map[string]any{
				"uri":  "spotify:track:abc",
				"name": "Song",
			},
		},
		"active_device_id": "device-1",
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/connect-state/v1/devices/hobs_") {
			return jsonResponse(http.StatusOK, statePayload), nil
		}
		return textResponse(http.StatusNotFound, "missing"), nil
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	_, err := VerifyPlayback(context.Background(), client, 10*time.Millisecond, 1*time.Millisecond)
	if err == nil {
		t.Fatalf("expected error")
	}
	var verifyErr *PlaybackVerifyError
	if !errors.As(err, &verifyErr) {
		t.Fatalf("expected PlaybackVerifyError, got %T", err)
	}
}

func TestVerifyPlaybackFailsOnMissingDeviceOrItem(t *testing.T) {
	statePayload := map[string]any{
		"devices":      map[string]any{},
		"player_state": map[string]any{"position_ms": 100},
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/connect-state/v1/devices/hobs_") {
			return jsonResponse(http.StatusOK, statePayload), nil
		}
		return textResponse(http.StatusNotFound, "missing"), nil
	})
	client := newConnectClientForTests(transport)
	client.session.connectDeviceID = "device"
	client.session.connectionID = "conn"
	client.session.registeredAt = time.Now()

	_, err := VerifyPlayback(context.Background(), client, 10*time.Millisecond, 1*time.Millisecond)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected missing in error, got %q", err.Error())
	}
}

func TestVerifyPlaybackNoopWhenTimeoutZero(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request")
		return nil, errors.New("unexpected request")
	})
	client := newConnectClientForTests(transport)
	_, err := VerifyPlayback(context.Background(), client, 0, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestPlaybackVerifyErrorUnwrap(t *testing.T) {
	base := errors.New("root")
	err := (&PlaybackVerifyError{LastErr: base}).Unwrap()
	if !errors.Is(err, base) {
		t.Fatalf("expected unwrap to return base error")
	}
	var nilErr *PlaybackVerifyError
	if nilErr.Unwrap() != nil {
		t.Fatalf("expected nil unwrap for nil receiver")
	}
}

func TestPlaybackVerifyErrorMessageIncludesMissingAndProgress(t *testing.T) {
	err := &PlaybackVerifyError{
		Timeout: time.Second,
		LastStatus: PlaybackStatus{
			ProgressMS: 0,
		},
	}
	msg := err.Error()
	if !strings.Contains(msg, "missing") || !strings.Contains(msg, "progress_ms") {
		t.Fatalf("unexpected message: %q", msg)
	}
}
