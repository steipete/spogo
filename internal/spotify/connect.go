package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/steipete/spogo/internal/cookies"
)

type ConnectOptions struct {
	Source   cookies.Source
	Market   string
	Language string
	Device   string
	Timeout  time.Duration
}

type ConnectClient struct {
	source       cookies.Source
	market       string
	language     string
	device       string
	client       *http.Client
	session      *connectSession
	hashes       *hashResolver
	webMu        sync.Mutex
	web          *Client
	searchURL    string
	searchClient *http.Client
}

func NewConnectClient(opts ConnectOptions) (*ConnectClient, error) {
	if opts.Source == nil {
		return nil, errors.New("cookie source required")
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	httpClient := &http.Client{Timeout: timeout}
	session := &connectSession{
		source: opts.Source,
		client: httpClient,
	}
	return &ConnectClient{
		source:   opts.Source,
		market:   opts.Market,
		language: opts.Language,
		device:   opts.Device,
		client:   httpClient,
		session:  session,
		hashes:   newHashResolver(httpClient, session),
	}, nil
}

func (c *ConnectClient) Search(ctx context.Context, kind, query string, limit, offset int) (SearchResult, error) {
	return c.search(ctx, kind, query, limit, offset)
}

func (c *ConnectClient) GetTrack(ctx context.Context, id string) (Item, error) {
	return c.trackInfo(ctx, id)
}

func (c *ConnectClient) GetAlbum(ctx context.Context, id string) (Item, error) {
	return c.albumInfo(ctx, id)
}

func (c *ConnectClient) GetArtist(ctx context.Context, id string) (Item, error) {
	return c.artistInfo(ctx, id)
}

func (c *ConnectClient) GetPlaylist(ctx context.Context, id string) (Item, error) {
	return c.playlistInfo(ctx, id)
}

func (c *ConnectClient) GetShow(ctx context.Context, id string) (Item, error) {
	return c.showInfo(ctx, id)
}

func (c *ConnectClient) GetEpisode(ctx context.Context, id string) (Item, error) {
	return c.episodeInfo(ctx, id)
}

func (c *ConnectClient) Playback(ctx context.Context) (PlaybackStatus, error) {
	return c.playback(ctx)
}

func (c *ConnectClient) Play(ctx context.Context, uri string) error {
	return c.play(ctx, uri)
}

func (c *ConnectClient) Pause(ctx context.Context) error {
	return c.pause(ctx)
}

func (c *ConnectClient) Next(ctx context.Context) error {
	return c.next(ctx)
}

func (c *ConnectClient) Previous(ctx context.Context) error {
	return c.previous(ctx)
}

func (c *ConnectClient) Seek(ctx context.Context, positionMS int) error {
	return c.seek(ctx, positionMS)
}

func (c *ConnectClient) Volume(ctx context.Context, volume int) error {
	return c.volume(ctx, volume)
}

func (c *ConnectClient) Shuffle(ctx context.Context, enabled bool) error {
	return c.shuffle(ctx, enabled)
}

func (c *ConnectClient) Repeat(ctx context.Context, mode string) error {
	return c.repeat(ctx, mode)
}

func (c *ConnectClient) Devices(ctx context.Context) ([]Device, error) {
	return c.devices(ctx)
}

func (c *ConnectClient) Transfer(ctx context.Context, deviceID string) error {
	return c.transfer(ctx, deviceID)
}

func (c *ConnectClient) QueueAdd(ctx context.Context, uri string) error {
	return c.queueAdd(ctx, uri)
}

func (c *ConnectClient) Queue(ctx context.Context) (Queue, error) {
	return c.queue(ctx)
}

func (c *ConnectClient) LibraryTracks(ctx context.Context, limit, offset int) ([]Item, int, error) {
	return nil, 0, fmt.Errorf("%w: library tracks not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) LibraryAlbums(ctx context.Context, limit, offset int) ([]Item, int, error) {
	return nil, 0, fmt.Errorf("%w: library albums not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) LibraryModify(ctx context.Context, path string, ids []string, method string) error {
	return fmt.Errorf("%w: library modify not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) FollowArtists(ctx context.Context, ids []string, method string) error {
	return fmt.Errorf("%w: follow artists not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) FollowedArtists(ctx context.Context, limit int, after string) ([]Item, int, string, error) {
	return nil, 0, "", fmt.Errorf("%w: followed artists not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) Playlists(ctx context.Context, limit, offset int) ([]Item, int, error) {
	return nil, 0, fmt.Errorf("%w: playlists not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) PlaylistTracks(ctx context.Context, id string, limit, offset int) ([]Item, int, error) {
	return nil, 0, fmt.Errorf("%w: playlist tracks not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) CreatePlaylist(ctx context.Context, name string, public, collaborative bool) (Item, error) {
	return Item{}, fmt.Errorf("%w: create playlist not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) AddTracks(ctx context.Context, playlistID string, uris []string) error {
	return fmt.Errorf("%w: add tracks not supported in connect engine yet", ErrUnsupported)
}

func (c *ConnectClient) RemoveTracks(ctx context.Context, playlistID string, uris []string) error {
	return fmt.Errorf("%w: remove tracks not supported in connect engine yet", ErrUnsupported)
}
