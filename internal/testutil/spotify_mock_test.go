package testutil

import (
	"context"
	"testing"

	"github.com/steipete/spogo/internal/spotify"
)

func TestSpotifyMockAllMethods(t *testing.T) {
	m := &SpotifyMock{
		SearchFn: func(context.Context, string, string, int, int) (spotify.SearchResult, error) {
			return spotify.SearchResult{}, nil
		},
		GetTrackFn:        func(context.Context, string) (spotify.Item, error) { return spotify.Item{}, nil },
		GetAlbumFn:        func(context.Context, string) (spotify.Item, error) { return spotify.Item{}, nil },
		GetArtistFn:       func(context.Context, string) (spotify.Item, error) { return spotify.Item{}, nil },
		GetPlaylistFn:     func(context.Context, string) (spotify.Item, error) { return spotify.Item{}, nil },
		GetShowFn:         func(context.Context, string) (spotify.Item, error) { return spotify.Item{}, nil },
		GetEpisodeFn:      func(context.Context, string) (spotify.Item, error) { return spotify.Item{}, nil },
		PlaybackFn:        func(context.Context) (spotify.PlaybackStatus, error) { return spotify.PlaybackStatus{}, nil },
		PlayFn:            func(context.Context, string) error { return nil },
		PauseFn:           func(context.Context) error { return nil },
		NextFn:            func(context.Context) error { return nil },
		PreviousFn:        func(context.Context) error { return nil },
		SeekFn:            func(context.Context, int) error { return nil },
		VolumeFn:          func(context.Context, int) error { return nil },
		ShuffleFn:         func(context.Context, bool) error { return nil },
		RepeatFn:          func(context.Context, string) error { return nil },
		DevicesFn:         func(context.Context) ([]spotify.Device, error) { return nil, nil },
		TransferFn:        func(context.Context, string) error { return nil },
		QueueAddFn:        func(context.Context, string) error { return nil },
		QueueFn:           func(context.Context) (spotify.Queue, error) { return spotify.Queue{}, nil },
		LibraryTracksFn:   func(context.Context, int, int) ([]spotify.Item, int, error) { return nil, 0, nil },
		LibraryAlbumsFn:   func(context.Context, int, int) ([]spotify.Item, int, error) { return nil, 0, nil },
		LibraryModifyFn:   func(context.Context, string, []string, string) error { return nil },
		FollowArtistsFn:   func(context.Context, []string, string) error { return nil },
		FollowedArtistsFn: func(context.Context, int, string) ([]spotify.Item, int, string, error) { return nil, 0, "", nil },
		PlaylistsFn:       func(context.Context, int, int) ([]spotify.Item, int, error) { return nil, 0, nil },
		PlaylistTracksFn:  func(context.Context, string, int, int) ([]spotify.Item, int, error) { return nil, 0, nil },
		CreatePlaylistFn:  func(context.Context, string, bool, bool) (spotify.Item, error) { return spotify.Item{}, nil },
		AddTracksFn:       func(context.Context, string, []string) error { return nil },
		RemoveTracksFn:    func(context.Context, string, []string) error { return nil },
	}
	_, _ = m.Search(context.Background(), "track", "q", 1, 0)
	_, _ = m.GetTrack(context.Background(), "1")
	_, _ = m.GetAlbum(context.Background(), "1")
	_, _ = m.GetArtist(context.Background(), "1")
	_, _ = m.GetPlaylist(context.Background(), "1")
	_, _ = m.GetShow(context.Background(), "1")
	_, _ = m.GetEpisode(context.Background(), "1")
	_, _ = m.Playback(context.Background())
	_ = m.Play(context.Background(), "uri")
	_ = m.Pause(context.Background())
	_ = m.Next(context.Background())
	_ = m.Previous(context.Background())
	_ = m.Seek(context.Background(), 1)
	_ = m.Volume(context.Background(), 1)
	_ = m.Shuffle(context.Background(), true)
	_ = m.Repeat(context.Background(), "off")
	_, _ = m.Devices(context.Background())
	_ = m.Transfer(context.Background(), "id")
	_ = m.QueueAdd(context.Background(), "uri")
	_, _ = m.Queue(context.Background())
	_, _, _ = m.LibraryTracks(context.Background(), 1, 0)
	_, _, _ = m.LibraryAlbums(context.Background(), 1, 0)
	_ = m.LibraryModify(context.Background(), "/me/tracks", []string{"1"}, "PUT")
	_ = m.FollowArtists(context.Background(), []string{"1"}, "PUT")
	_, _, _, _ = m.FollowedArtists(context.Background(), 1, "")
	_, _, _ = m.Playlists(context.Background(), 1, 0)
	_, _, _ = m.PlaylistTracks(context.Background(), "1", 1, 0)
	_, _ = m.CreatePlaylist(context.Background(), "name", true, false)
	_ = m.AddTracks(context.Background(), "p", []string{"u"})
	_ = m.RemoveTracks(context.Background(), "p", []string{"u"})
}

func TestSpotifyMockNotImplemented(t *testing.T) {
	m := &SpotifyMock{}
	if _, err := m.Search(context.Background(), "track", "q", 1, 0); err == nil {
		t.Fatalf("expected error")
	}
	if err := m.Play(context.Background(), "uri"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := m.GetTrack(context.Background(), "1"); err == nil {
		t.Fatalf("expected error")
	}
	if err := m.Pause(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := m.Queue(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}
