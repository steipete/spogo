package testutil

import (
	"context"
	"errors"

	"github.com/steipete/spogo/internal/spotify"
)

var ErrNotImplemented = errors.New("not implemented")

type SpotifyMock struct {
	SearchFn          func(context.Context, string, string, int, int) (spotify.SearchResult, error)
	GetTrackFn        func(context.Context, string) (spotify.Item, error)
	GetAlbumFn        func(context.Context, string) (spotify.Item, error)
	GetArtistFn       func(context.Context, string) (spotify.Item, error)
	GetPlaylistFn     func(context.Context, string) (spotify.Item, error)
	GetShowFn         func(context.Context, string) (spotify.Item, error)
	GetEpisodeFn      func(context.Context, string) (spotify.Item, error)
	PlaybackFn        func(context.Context) (spotify.PlaybackStatus, error)
	PlayFn            func(context.Context, string) error
	PauseFn           func(context.Context) error
	NextFn            func(context.Context) error
	PreviousFn        func(context.Context) error
	SeekFn            func(context.Context, int) error
	VolumeFn          func(context.Context, int) error
	ShuffleFn         func(context.Context, bool) error
	RepeatFn          func(context.Context, string) error
	DevicesFn         func(context.Context) ([]spotify.Device, error)
	TransferFn        func(context.Context, string) error
	QueueAddFn        func(context.Context, string) error
	QueueFn           func(context.Context) (spotify.Queue, error)
	LibraryTracksFn   func(context.Context, int, int) ([]spotify.Item, int, error)
	LibraryAlbumsFn   func(context.Context, int, int) ([]spotify.Item, int, error)
	LibraryModifyFn   func(context.Context, string, []string, string) error
	FollowArtistsFn   func(context.Context, []string, string) error
	FollowedArtistsFn func(context.Context, int, string) ([]spotify.Item, int, string, error)
	PlaylistsFn       func(context.Context, int, int) ([]spotify.Item, int, error)
	PlaylistTracksFn  func(context.Context, string, int, int) ([]spotify.Item, int, error)
	CreatePlaylistFn  func(context.Context, string, bool, bool) (spotify.Item, error)
	AddTracksFn       func(context.Context, string, []string) error
	RemoveTracksFn    func(context.Context, string, []string) error
}

func (m *SpotifyMock) Search(ctx context.Context, kind, query string, limit, offset int) (spotify.SearchResult, error) {
	if m.SearchFn == nil {
		return spotify.SearchResult{}, ErrNotImplemented
	}
	return m.SearchFn(ctx, kind, query, limit, offset)
}

func (m *SpotifyMock) GetTrack(ctx context.Context, id string) (spotify.Item, error) {
	if m.GetTrackFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.GetTrackFn(ctx, id)
}

func (m *SpotifyMock) GetAlbum(ctx context.Context, id string) (spotify.Item, error) {
	if m.GetAlbumFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.GetAlbumFn(ctx, id)
}

func (m *SpotifyMock) GetArtist(ctx context.Context, id string) (spotify.Item, error) {
	if m.GetArtistFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.GetArtistFn(ctx, id)
}

func (m *SpotifyMock) GetPlaylist(ctx context.Context, id string) (spotify.Item, error) {
	if m.GetPlaylistFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.GetPlaylistFn(ctx, id)
}

func (m *SpotifyMock) GetShow(ctx context.Context, id string) (spotify.Item, error) {
	if m.GetShowFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.GetShowFn(ctx, id)
}

func (m *SpotifyMock) GetEpisode(ctx context.Context, id string) (spotify.Item, error) {
	if m.GetEpisodeFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.GetEpisodeFn(ctx, id)
}

func (m *SpotifyMock) Playback(ctx context.Context) (spotify.PlaybackStatus, error) {
	if m.PlaybackFn == nil {
		return spotify.PlaybackStatus{}, ErrNotImplemented
	}
	return m.PlaybackFn(ctx)
}

func (m *SpotifyMock) Play(ctx context.Context, uri string) error {
	if m.PlayFn == nil {
		return ErrNotImplemented
	}
	return m.PlayFn(ctx, uri)
}

func (m *SpotifyMock) Pause(ctx context.Context) error {
	if m.PauseFn == nil {
		return ErrNotImplemented
	}
	return m.PauseFn(ctx)
}

func (m *SpotifyMock) Next(ctx context.Context) error {
	if m.NextFn == nil {
		return ErrNotImplemented
	}
	return m.NextFn(ctx)
}

func (m *SpotifyMock) Previous(ctx context.Context) error {
	if m.PreviousFn == nil {
		return ErrNotImplemented
	}
	return m.PreviousFn(ctx)
}

func (m *SpotifyMock) Seek(ctx context.Context, positionMS int) error {
	if m.SeekFn == nil {
		return ErrNotImplemented
	}
	return m.SeekFn(ctx, positionMS)
}

func (m *SpotifyMock) Volume(ctx context.Context, volume int) error {
	if m.VolumeFn == nil {
		return ErrNotImplemented
	}
	return m.VolumeFn(ctx, volume)
}

func (m *SpotifyMock) Shuffle(ctx context.Context, enabled bool) error {
	if m.ShuffleFn == nil {
		return ErrNotImplemented
	}
	return m.ShuffleFn(ctx, enabled)
}

func (m *SpotifyMock) Repeat(ctx context.Context, mode string) error {
	if m.RepeatFn == nil {
		return ErrNotImplemented
	}
	return m.RepeatFn(ctx, mode)
}

func (m *SpotifyMock) Devices(ctx context.Context) ([]spotify.Device, error) {
	if m.DevicesFn == nil {
		return nil, ErrNotImplemented
	}
	return m.DevicesFn(ctx)
}

func (m *SpotifyMock) Transfer(ctx context.Context, deviceID string) error {
	if m.TransferFn == nil {
		return ErrNotImplemented
	}
	return m.TransferFn(ctx, deviceID)
}

func (m *SpotifyMock) QueueAdd(ctx context.Context, uri string) error {
	if m.QueueAddFn == nil {
		return ErrNotImplemented
	}
	return m.QueueAddFn(ctx, uri)
}

func (m *SpotifyMock) Queue(ctx context.Context) (spotify.Queue, error) {
	if m.QueueFn == nil {
		return spotify.Queue{}, ErrNotImplemented
	}
	return m.QueueFn(ctx)
}

func (m *SpotifyMock) LibraryTracks(ctx context.Context, limit, offset int) ([]spotify.Item, int, error) {
	if m.LibraryTracksFn == nil {
		return nil, 0, ErrNotImplemented
	}
	return m.LibraryTracksFn(ctx, limit, offset)
}

func (m *SpotifyMock) LibraryAlbums(ctx context.Context, limit, offset int) ([]spotify.Item, int, error) {
	if m.LibraryAlbumsFn == nil {
		return nil, 0, ErrNotImplemented
	}
	return m.LibraryAlbumsFn(ctx, limit, offset)
}

func (m *SpotifyMock) LibraryModify(ctx context.Context, path string, ids []string, method string) error {
	if m.LibraryModifyFn == nil {
		return ErrNotImplemented
	}
	return m.LibraryModifyFn(ctx, path, ids, method)
}

func (m *SpotifyMock) FollowArtists(ctx context.Context, ids []string, method string) error {
	if m.FollowArtistsFn == nil {
		return ErrNotImplemented
	}
	return m.FollowArtistsFn(ctx, ids, method)
}

func (m *SpotifyMock) FollowedArtists(ctx context.Context, limit int, after string) ([]spotify.Item, int, string, error) {
	if m.FollowedArtistsFn == nil {
		return nil, 0, "", ErrNotImplemented
	}
	return m.FollowedArtistsFn(ctx, limit, after)
}

func (m *SpotifyMock) Playlists(ctx context.Context, limit, offset int) ([]spotify.Item, int, error) {
	if m.PlaylistsFn == nil {
		return nil, 0, ErrNotImplemented
	}
	return m.PlaylistsFn(ctx, limit, offset)
}

func (m *SpotifyMock) PlaylistTracks(ctx context.Context, id string, limit, offset int) ([]spotify.Item, int, error) {
	if m.PlaylistTracksFn == nil {
		return nil, 0, ErrNotImplemented
	}
	return m.PlaylistTracksFn(ctx, id, limit, offset)
}

func (m *SpotifyMock) CreatePlaylist(ctx context.Context, name string, public, collaborative bool) (spotify.Item, error) {
	if m.CreatePlaylistFn == nil {
		return spotify.Item{}, ErrNotImplemented
	}
	return m.CreatePlaylistFn(ctx, name, public, collaborative)
}

func (m *SpotifyMock) AddTracks(ctx context.Context, playlistID string, uris []string) error {
	if m.AddTracksFn == nil {
		return ErrNotImplemented
	}
	return m.AddTracksFn(ctx, playlistID, uris)
}

func (m *SpotifyMock) RemoveTracks(ctx context.Context, playlistID string, uris []string) error {
	if m.RemoveTracksFn == nil {
		return ErrNotImplemented
	}
	return m.RemoveTracksFn(ctx, playlistID, uris)
}
