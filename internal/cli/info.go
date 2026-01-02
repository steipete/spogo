package cli

import (
	"context"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/spotify"
)

type TrackCmd struct {
	Info InfoTrackCmd `kong:"cmd,help='Track info.'"`
}

type AlbumCmd struct {
	Info InfoAlbumCmd `kong:"cmd,help='Album info.'"`
}

type ArtistCmd struct {
	Info InfoArtistCmd `kong:"cmd,help='Artist info.'"`
}

type PlaylistCmd struct {
	Info   InfoPlaylistCmd   `kong:"cmd,help='Playlist info.'"`
	Create PlaylistCreateCmd `kong:"cmd,help='Create playlist.'"`
	Add    PlaylistAddCmd    `kong:"cmd,help='Add tracks to playlist.'"`
	Remove PlaylistRemoveCmd `kong:"cmd,help='Remove tracks from playlist.'"`
	Tracks PlaylistTracksCmd `kong:"cmd,help='List playlist tracks.'"`
}

type ShowCmd struct {
	Info InfoShowCmd `kong:"cmd,help='Show info.'"`
}

type EpisodeCmd struct {
	Info InfoEpisodeCmd `kong:"cmd,help='Episode info.'"`
}

type InfoArgs struct {
	ID string `arg:"" required:"" help:"Spotify ID, URI, or URL."`
}

type InfoTrackCmd struct{ InfoArgs }

type InfoAlbumCmd struct{ InfoArgs }

type InfoArtistCmd struct{ InfoArgs }

type InfoPlaylistCmd struct{ InfoArgs }

type InfoShowCmd struct{ InfoArgs }

type InfoEpisodeCmd struct{ InfoArgs }

func (cmd *InfoTrackCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.ID, "track")
	if err != nil {
		return err
	}
	item, err := client.GetTrack(context.Background(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Output.Emit(item, []string{itemPlain(item)}, []string{itemHuman(ctx.Output, item)})
}

func (cmd *InfoAlbumCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.ID, "album")
	if err != nil {
		return err
	}
	item, err := client.GetAlbum(context.Background(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Output.Emit(item, []string{itemPlain(item)}, []string{itemHuman(ctx.Output, item)})
}

func (cmd *InfoArtistCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.ID, "artist")
	if err != nil {
		return err
	}
	item, err := client.GetArtist(context.Background(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Output.Emit(item, []string{itemPlain(item)}, []string{itemHuman(ctx.Output, item)})
}

func (cmd *InfoPlaylistCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.ID, "playlist")
	if err != nil {
		return err
	}
	item, err := client.GetPlaylist(context.Background(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Output.Emit(item, []string{itemPlain(item)}, []string{itemHuman(ctx.Output, item)})
}

func (cmd *InfoShowCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.ID, "show")
	if err != nil {
		return err
	}
	item, err := client.GetShow(context.Background(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Output.Emit(item, []string{itemPlain(item)}, []string{itemHuman(ctx.Output, item)})
}

func (cmd *InfoEpisodeCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.ID, "episode")
	if err != nil {
		return err
	}
	item, err := client.GetEpisode(context.Background(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Output.Emit(item, []string{itemPlain(item)}, []string{itemHuman(ctx.Output, item)})
}
