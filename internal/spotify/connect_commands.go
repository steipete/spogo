package spotify

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

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
	var payload map[string]any
	if isContextURI(uri) {
		payload = map[string]any{
			"command": map[string]any{
				"endpoint": "play",
				"logging_params": map[string]any{
					"command_id": randomHex(32),
				},
				"context": map[string]any{
					"uri": uri,
					"url": "context://" + uri,
				},
			},
		}
	} else {
		payload = map[string]any{
			"command": map[string]any{
				"endpoint": "play",
				"logging_params": map[string]any{
					"command_id": randomHex(32),
				},
				"context": map[string]any{
					"uri": uri,
					"url": "context://" + uri,
				},
				"options": map[string]any{
					"skip_to": map[string]any{
						"track_uri": uri,
					},
				},
			},
		}
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
