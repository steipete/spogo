package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/config"
	"github.com/steipete/spogo/internal/spotify"
)

type DeviceCmd struct {
	List  DeviceListCmd  `kong:"cmd,help='List devices.'"`
	Set   DeviceSetCmd   `kong:"cmd,help='Set active device.'"`
	Show  DeviceShowCmd  `kong:"cmd,help='Show active device and configured target.'"`
	Clear DeviceClearCmd `kong:"cmd,help='Clear saved device for current profile.'"`
}

type DeviceListCmd struct{}

type DeviceSetCmd struct {
	Device string `arg:"" required:"" help:"Device name or id (supports unique partial match)."`
	Save   bool   `help:"Persist device selection in the current profile config."`
}

type DeviceShowCmd struct{}

type DeviceClearCmd struct{}

func (cmd *DeviceListCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	devices, err := client.Devices(context.Background())
	if err != nil {
		return err
	}
	sort.SliceStable(devices, func(i, j int) bool {
		if devices[i].Active == devices[j].Active {
			return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
		}
		return devices[i].Active
	})
	plain := make([]string, 0, len(devices))
	human := make([]string, 0, len(devices))
	for _, device := range devices {
		plain = append(plain, fmt.Sprintf("%s\t%s\t%t", device.ID, device.Name, device.Active))
		label := device.Name
		if device.Active {
			label = ctx.Output.Theme.Accent(label)
		}
		human = append(human, fmt.Sprintf("%s (%s) %s", label, device.Type, strings.TrimSpace(activeMarker(device.Active))))
	}
	return ctx.Output.Emit(devices, plain, human)
}

func (cmd *DeviceSetCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	devices, err := client.Devices(context.Background())
	if err != nil {
		return err
	}
	id, err := spotify.ResolveDeviceID(devices, cmd.Device)
	if err != nil {
		return err
	}
	if err := client.Transfer(context.Background(), id); err != nil {
		return err
	}
	if cmd.Save {
		if ctx.Config == nil {
			return errors.New("config not loaded")
		}
		profile := ctx.Profile
		profile.Device = id
		ctx.Profile = profile
		ctx.Config.SetProfile(ctx.ProfileKey, profile)
		if err := config.Save(ctx.ConfigPath, ctx.Config); err != nil {
			return err
		}
	}
	return ctx.Output.Emit(map[string]any{"status": "ok", "device": id}, []string{"ok"}, []string{fmt.Sprintf("Switched to %s", id)})
}

func (cmd *DeviceShowCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	devices, err := client.Devices(context.Background())
	if err != nil {
		return err
	}

	var active spotify.Device
	for _, d := range devices {
		if d.Active {
			active = d
			break
		}
	}

	selector := strings.TrimSpace(ctx.Profile.Device)
	targetID := ""
	var target spotify.Device
	if selector != "" {
		id, rerr := spotify.ResolveDeviceID(devices, selector)
		if rerr != nil {
			return rerr
		}
		targetID = id
		for _, d := range devices {
			if d.ID == targetID {
				target = d
				break
			}
		}
	}

	payload := map[string]any{
		"status": "ok",
		"active": active,
		"target": target,
	}
	plain := []string{fmt.Sprintf("%s\t%s\t%s\t%s", active.ID, active.Name, targetID, target.Name)}

	activeLabel := "(none)"
	if strings.TrimSpace(active.ID) != "" || strings.TrimSpace(active.Name) != "" {
		activeLabel = fmt.Sprintf("%s (%s)", active.Name, active.ID)
	}
	targetLabel := "(none)"
	if selector != "" {
		targetLabel = fmt.Sprintf("%s (%s)", target.Name, targetID)
	}
	human := []string{
		fmt.Sprintf("Active: %s", activeLabel),
		fmt.Sprintf("Target: %s", targetLabel),
	}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *DeviceClearCmd) Run(ctx *app.Context) error {
	if ctx.Config == nil {
		return errors.New("config not loaded")
	}
	profile := ctx.Profile
	profile.Device = ""
	ctx.Profile = profile
	ctx.Config.SetProfile(ctx.ProfileKey, profile)
	if err := config.Save(ctx.ConfigPath, ctx.Config); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{"Cleared saved device"})
}

func activeMarker(active bool) string {
	if active {
		return "(active)"
	}
	return ""
}
