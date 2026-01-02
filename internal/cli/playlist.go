package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
)

type PlaylistCreateCmd struct {
	Name   string `arg:"" required:"" help:"Playlist name."`
	Public bool   `help:"Create public playlist."`
	Collab bool   `help:"Create collaborative playlist."`
}

type PlaylistAddCmd struct {
	Playlist string   `arg:"" required:"" help:"Playlist ID/URL/URI."`
	Tracks   []string `arg:"" required:"" help:"Track IDs/URLs/URIs."`
}

type PlaylistRemoveCmd struct {
	Playlist string   `arg:"" required:"" help:"Playlist ID/URL/URI."`
	Tracks   []string `arg:"" required:"" help:"Track IDs/URLs/URIs."`
}

type PlaylistTracksCmd struct {
	Playlist string `arg:"" required:"" help:"Playlist ID/URL/URI."`
	Limit    int    `help:"Limit results." default:"50"`
	Offset   int    `help:"Offset results." default:"0"`
}

func (cmd *PlaylistCreateCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	item, err := client.CreatePlaylist(context.Background(), cmd.Name, cmd.Public, cmd.Collab)
	if err != nil {
		return err
	}
	plain := []string{itemPlain(item)}
	human := []string{fmt.Sprintf("Created %s", itemHuman(ctx.Output, item))}
	return ctx.Output.Emit(item, plain, human)
}

func (cmd *PlaylistAddCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	playlist, err := spotify.ParseTypedID(cmd.Playlist, "playlist")
	if err != nil {
		return err
	}
	uris, err := trackURIs(cmd.Tracks)
	if err != nil {
		return err
	}
	if err := client.AddTracks(context.Background(), playlist.ID, uris); err != nil {
		return err
	}
	plain := []string{"ok"}
	human := []string{fmt.Sprintf("Added %d tracks", len(uris))}
	return ctx.Output.Emit(map[string]any{"status": "ok", "count": len(uris)}, plain, human)
}

func (cmd *PlaylistRemoveCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	playlist, err := spotify.ParseTypedID(cmd.Playlist, "playlist")
	if err != nil {
		return err
	}
	uris, err := trackURIs(cmd.Tracks)
	if err != nil {
		return err
	}
	if err := client.RemoveTracks(context.Background(), playlist.ID, uris); err != nil {
		return err
	}
	plain := []string{"ok"}
	human := []string{fmt.Sprintf("Removed %d tracks", len(uris))}
	return ctx.Output.Emit(map[string]any{"status": "ok", "count": len(uris)}, plain, human)
}

func (cmd *PlaylistTracksCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	playlist, err := spotify.ParseTypedID(cmd.Playlist, "playlist")
	if err != nil {
		return err
	}
	limit := clampLimit(cmd.Limit)
	items, total, err := client.PlaylistTracks(context.Background(), playlist.ID, limit, cmd.Offset)
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, items)
	if ctx.Output.Format == output.FormatHuman {
		human = append([]string{fmt.Sprintf("Tracks: %d", total)}, human...)
	}
	payload := map[string]any{"total": total, "items": items}
	return ctx.Output.Emit(payload, plain, human)
}

func trackURIs(inputs []string) ([]string, error) {
	uris := make([]string, 0, len(inputs))
	for _, input := range inputs {
		res, err := spotify.ParseTypedID(strings.TrimSpace(input), "track")
		if err != nil {
			return nil, err
		}
		if res.URI == "" {
			return nil, fmt.Errorf("invalid track input")
		}
		uris = append(uris, res.URI)
	}
	return uris, nil
}
