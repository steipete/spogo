package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/spotify"
)

type LibraryCmd struct {
	Tracks    LibraryTracksCmd    `kong:"cmd,help='Track library.'"`
	Albums    LibraryAlbumsCmd    `kong:"cmd,help='Album library.'"`
	Artists   LibraryArtistsCmd   `kong:"cmd,help='Artist library.'"`
	Playlists LibraryPlaylistsCmd `kong:"cmd,help='Playlist library.'"`
}

type LibraryTracksCmd struct {
	List   LibraryTracksListCmd   `kong:"cmd,help='List saved tracks.'"`
	Add    LibraryTracksAddCmd    `kong:"cmd,help='Save tracks.'"`
	Remove LibraryTracksRemoveCmd `kong:"cmd,help='Remove saved tracks.'"`
}

type LibraryAlbumsCmd struct {
	List   LibraryAlbumsListCmd   `kong:"cmd,help='List saved albums.'"`
	Add    LibraryAlbumsAddCmd    `kong:"cmd,help='Save albums.'"`
	Remove LibraryAlbumsRemoveCmd `kong:"cmd,help='Remove saved albums.'"`
}

type LibraryArtistsCmd struct {
	List     LibraryArtistsListCmd     `kong:"cmd,help='List followed artists.'"`
	Follow   LibraryArtistsFollowCmd   `kong:"cmd,help='Follow artists.'"`
	Unfollow LibraryArtistsUnfollowCmd `kong:"cmd,help='Unfollow artists.'"`
}

type LibraryPlaylistsCmd struct {
	List LibraryPlaylistsListCmd `kong:"cmd,help='List playlists.'"`
}

type LibraryTracksListCmd struct {
	Limit  int `help:"Limit results." default:"50"`
	Offset int `help:"Offset results." default:"0"`
}

type LibraryTracksAddCmd struct {
	IDs []string `arg:"" required:"" help:"Track IDs/URLs/URIs."`
}

type LibraryTracksRemoveCmd struct {
	IDs []string `arg:"" required:"" help:"Track IDs/URLs/URIs."`
}

type LibraryAlbumsListCmd struct {
	Limit  int `help:"Limit results." default:"50"`
	Offset int `help:"Offset results." default:"0"`
}

type LibraryAlbumsAddCmd struct {
	IDs []string `arg:"" required:"" help:"Album IDs/URLs/URIs."`
}

type LibraryAlbumsRemoveCmd struct {
	IDs []string `arg:"" required:"" help:"Album IDs/URLs/URIs."`
}

type LibraryArtistsListCmd struct {
	Limit  int    `help:"Limit results." default:"50"`
	After  string `help:"Artist ID to start after (pagination)."`
	Offset int    `help:"Offset results (not supported by Spotify)."`
}

type LibraryArtistsFollowCmd struct {
	IDs []string `arg:"" required:"" help:"Artist IDs/URLs/URIs."`
}

type LibraryArtistsUnfollowCmd struct {
	IDs []string `arg:"" required:"" help:"Artist IDs/URLs/URIs."`
}

type LibraryPlaylistsListCmd struct {
	Limit  int `help:"Limit results." default:"50"`
	Offset int `help:"Offset results." default:"0"`
}

func (cmd *LibraryTracksListCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	limit := clampLimit(cmd.Limit)
	items, total, err := client.LibraryTracks(context.Background(), limit, cmd.Offset)
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, items)
	payload := map[string]any{"total": total, "items": items}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *LibraryTracksAddCmd) Run(ctx *app.Context) error {
	ids, err := parseIDs(cmd.IDs, "track")
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.LibraryModify(context.Background(), "/me/tracks", ids, "PUT"); err != nil {
		return err
	}
	return ok(ctx, len(ids))
}

func (cmd *LibraryTracksRemoveCmd) Run(ctx *app.Context) error {
	ids, err := parseIDs(cmd.IDs, "track")
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.LibraryModify(context.Background(), "/me/tracks", ids, "DELETE"); err != nil {
		return err
	}
	return ok(ctx, len(ids))
}

func (cmd *LibraryAlbumsListCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	limit := clampLimit(cmd.Limit)
	items, total, err := client.LibraryAlbums(context.Background(), limit, cmd.Offset)
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, items)
	payload := map[string]any{"total": total, "items": items}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *LibraryAlbumsAddCmd) Run(ctx *app.Context) error {
	ids, err := parseIDs(cmd.IDs, "album")
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.LibraryModify(context.Background(), "/me/albums", ids, "PUT"); err != nil {
		return err
	}
	return ok(ctx, len(ids))
}

func (cmd *LibraryAlbumsRemoveCmd) Run(ctx *app.Context) error {
	ids, err := parseIDs(cmd.IDs, "album")
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.LibraryModify(context.Background(), "/me/albums", ids, "DELETE"); err != nil {
		return err
	}
	return ok(ctx, len(ids))
}

func (cmd *LibraryArtistsListCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if cmd.Offset != 0 {
		return fmt.Errorf("offset not supported; use --after with an artist id")
	}
	limit := clampLimit(cmd.Limit)
	items, total, next, err := client.FollowedArtists(context.Background(), limit, cmd.After)
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, items)
	payload := map[string]any{"total": total, "items": items, "next_after": next}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *LibraryArtistsFollowCmd) Run(ctx *app.Context) error {
	ids, err := parseIDs(cmd.IDs, "artist")
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.FollowArtists(context.Background(), ids, "PUT"); err != nil {
		return err
	}
	return ok(ctx, len(ids))
}

func (cmd *LibraryArtistsUnfollowCmd) Run(ctx *app.Context) error {
	ids, err := parseIDs(cmd.IDs, "artist")
	if err != nil {
		return err
	}
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	if err := client.FollowArtists(context.Background(), ids, "DELETE"); err != nil {
		return err
	}
	return ok(ctx, len(ids))
}

func (cmd *LibraryPlaylistsListCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	limit := clampLimit(cmd.Limit)
	items, total, err := client.Playlists(context.Background(), limit, cmd.Offset)
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, items)
	payload := map[string]any{"total": total, "items": items}
	return ctx.Output.Emit(payload, plain, human)
}

func parseIDs(inputs []string, kind string) ([]string, error) {
	ids := make([]string, 0, len(inputs))
	for _, input := range inputs {
		res, err := spotify.ParseTypedID(strings.TrimSpace(input), kind)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.ID)
	}
	return ids, nil
}

func ok(ctx *app.Context, count int) error {
	payload := map[string]any{"status": "ok", "count": count}
	plain := []string{"ok"}
	human := []string{fmt.Sprintf("Updated %d items", count)}
	return ctx.Output.Emit(payload, plain, human)
}
