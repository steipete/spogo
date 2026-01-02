package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
)

type SearchCmd struct {
	Track    SearchTrackCmd    `kong:"cmd,help='Search tracks.'"`
	Album    SearchAlbumCmd    `kong:"cmd,help='Search albums.'"`
	Artist   SearchArtistCmd   `kong:"cmd,help='Search artists.'"`
	Playlist SearchPlaylistCmd `kong:"cmd,help='Search playlists.'"`
	Episode  SearchEpisodeCmd  `kong:"cmd,help='Search episodes.'"`
	Show     SearchShowCmd     `kong:"cmd,help='Search shows.'"`
}

type SearchArgs struct {
	Query  string `arg:"" required:"" help:"Search query."`
	Limit  int    `help:"Limit results." default:"20"`
	Offset int    `help:"Offset results." default:"0"`
}

type SearchTrackCmd struct{ SearchArgs }

type SearchAlbumCmd struct{ SearchArgs }

type SearchArtistCmd struct{ SearchArgs }

type SearchPlaylistCmd struct{ SearchArgs }

type SearchEpisodeCmd struct{ SearchArgs }

type SearchShowCmd struct{ SearchArgs }

func (cmd *SearchTrackCmd) Run(ctx *app.Context) error {
	return runSearch(ctx, "track", cmd.SearchArgs)
}

func (cmd *SearchAlbumCmd) Run(ctx *app.Context) error {
	return runSearch(ctx, "album", cmd.SearchArgs)
}

func (cmd *SearchArtistCmd) Run(ctx *app.Context) error {
	return runSearch(ctx, "artist", cmd.SearchArgs)
}

func (cmd *SearchPlaylistCmd) Run(ctx *app.Context) error {
	return runSearch(ctx, "playlist", cmd.SearchArgs)
}

func (cmd *SearchEpisodeCmd) Run(ctx *app.Context) error {
	return runSearch(ctx, "episode", cmd.SearchArgs)
}

func (cmd *SearchShowCmd) Run(ctx *app.Context) error {
	return runSearch(ctx, "show", cmd.SearchArgs)
}

func runSearch(ctx *app.Context, kind string, args SearchArgs) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	limit := clampLimit(args.Limit)
	if args.Limit != limit {
		ctx.Output.Errorf("limit capped at %d", limit)
	}
	res, err := client.Search(context.Background(), kind, args.Query, limit, args.Offset)
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, res.Items)
	header := fmt.Sprintf("%s results: %d", strings.ToUpper(kind), res.Total)
	if ctx.Output.Format == output.FormatHuman {
		human = append([]string{header}, human...)
	}
	return ctx.Output.Emit(res, plain, human)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 50 {
		return 50
	}
	return limit
}
