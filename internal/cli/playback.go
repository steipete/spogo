package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/spotify"
)

type PlayCmd struct {
	Item string `arg:"" optional:"" help:"Spotify ID/URL/URI."`
	Type string `help:"Type for raw IDs (track|album|playlist|show|episode)."`
}

type PauseCmd struct{}

type NextCmd struct{}

type PrevCmd struct{}

type SeekCmd struct {
	Position string `arg:"" required:"" help:"Position ms or mm:ss."`
}

type VolumeCmd struct {
	Level int `arg:"" required:"" help:"Volume 0-100."`
}

type ShuffleCmd struct {
	State string `arg:"" required:"" help:"on|off."`
}

type RepeatCmd struct {
	Mode string `arg:"" required:"" help:"off|track|context."`
}

type StatusCmd struct{}

func (cmd *PlayCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	uri := ""
	if cmd.Item != "" {
		res, err := spotify.ParseResource(cmd.Item)
		if err != nil {
			return err
		}
		if res.URI == "" {
			if cmd.Type == "" {
				return errors.New("type required for raw id")
			}
			res.Type = cmd.Type
			res.URI = "spotify:" + cmd.Type + ":" + res.ID
		}
		uri = res.URI
	}
	if err := client.Play(context.Background(), uri); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{"Playback started"})
}

func (cmd *PauseCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.Pause(context.Background()); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{"Playback paused"})
}

func (cmd *NextCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.Next(context.Background()); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{"Skipped to next"})
}

func (cmd *PrevCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.Previous(context.Background()); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{"Skipped to previous"})
}

func (cmd *SeekCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	position, err := parsePosition(cmd.Position)
	if err != nil {
		return err
	}
	if err := client.Seek(context.Background(), position); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]any{"status": "ok", "position_ms": position}, []string{"ok"}, []string{fmt.Sprintf("Seeked to %s", humanDuration(position))})
}

func (cmd *VolumeCmd) Run(ctx *app.Context) error {
	if cmd.Level < 0 || cmd.Level > 100 {
		return fmt.Errorf("volume must be 0-100")
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.Volume(context.Background(), cmd.Level); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]any{"status": "ok", "volume": cmd.Level}, []string{"ok"}, []string{fmt.Sprintf("Volume %d", cmd.Level)})
}

func (cmd *ShuffleCmd) Run(ctx *app.Context) error {
	state, err := parseToggle(cmd.State)
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.Shuffle(context.Background(), state); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]any{"status": "ok", "shuffle": state}, []string{"ok"}, []string{fmt.Sprintf("Shuffle %s", onOff(state))})
}

func (cmd *RepeatCmd) Run(ctx *app.Context) error {
	mode := strings.ToLower(strings.TrimSpace(cmd.Mode))
	if mode != "off" && mode != "track" && mode != "context" {
		return fmt.Errorf("repeat must be off|track|context")
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.Repeat(context.Background(), mode); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]any{"status": "ok", "repeat": mode}, []string{"ok"}, []string{fmt.Sprintf("Repeat %s", mode)})
}

func (cmd *StatusCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	status, err := client.Playback(context.Background())
	if err != nil {
		return err
	}
	plain := []string{playbackPlain(status)}
	human := []string{playbackHuman(ctx.Output, status)}
	return ctx.Output.Emit(status, plain, human)
}

func parsePosition(input string) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("position required")
	}
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid position format")
		}
		min, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		sec, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		d := time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
		return int(d / time.Millisecond), nil
	}
	ms, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}
	return ms, nil
}

func parseToggle(input string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "on", "true", "1", "yes":
		return true, nil
	case "off", "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("expected on|off")
	}
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
