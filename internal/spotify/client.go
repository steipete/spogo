package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Options struct {
	TokenProvider TokenProvider
	HTTPClient    *http.Client
	BaseURL       string
	Market        string
	Language      string
	Device        string
	Timeout       time.Duration
}

type Client struct {
	baseURL  string
	market   string
	language string
	device   string
	client   *http.Client
	provider TokenProvider

	mu        sync.Mutex
	lastToken Token
}

func NewClient(opts Options) (*Client, error) {
	if opts.TokenProvider == nil {
		return nil, errors.New("token provider required")
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://api.spotify.com/v1"
	}
	client := opts.HTTPClient
	if client == nil {
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = 10 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}
	return &Client{
		baseURL:  baseURL,
		market:   opts.Market,
		language: opts.Language,
		device:   opts.Device,
		client:   client,
		provider: opts.TokenProvider,
	}, nil
}

func (c *Client) Search(ctx context.Context, kind, query string, limit, offset int) (SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("type", kind)
	params.Set("limit", fmt.Sprint(limit))
	params.Set("offset", fmt.Sprint(offset))
	var response map[string]searchContainer
	if err := c.get(ctx, "/search", params, &response); err != nil {
		return SearchResult{}, err
	}
	container, ok := response[kind]
	if !ok {
		return SearchResult{}, fmt.Errorf("missing %s result", kind)
	}
	items := make([]Item, 0, len(container.Items))
	for _, raw := range container.Items {
		item, err := mapSearchItem(kind, raw)
		if err != nil {
			return SearchResult{}, err
		}
		items = append(items, item)
	}
	return SearchResult{
		Type:   kind,
		Limit:  container.Limit,
		Offset: container.Offset,
		Total:  container.Total,
		Items:  items,
	}, nil
}

func (c *Client) GetTrack(ctx context.Context, id string) (Item, error) {
	var raw trackItem
	if err := c.get(ctx, "/tracks/"+id, url.Values{}, &raw); err != nil {
		return Item{}, err
	}
	return mapTrack(raw), nil
}

func (c *Client) GetAlbum(ctx context.Context, id string) (Item, error) {
	var raw albumItem
	if err := c.get(ctx, "/albums/"+id, url.Values{}, &raw); err != nil {
		return Item{}, err
	}
	return mapAlbum(raw), nil
}

func (c *Client) GetArtist(ctx context.Context, id string) (Item, error) {
	var raw artistItem
	if err := c.get(ctx, "/artists/"+id, nil, &raw); err != nil {
		return Item{}, err
	}
	return mapArtist(raw), nil
}

func (c *Client) GetPlaylist(ctx context.Context, id string) (Item, error) {
	var raw playlistItem
	if err := c.get(ctx, "/playlists/"+id, nil, &raw); err != nil {
		return Item{}, err
	}
	return mapPlaylist(raw), nil
}

func (c *Client) GetShow(ctx context.Context, id string) (Item, error) {
	var raw showItem
	if err := c.get(ctx, "/shows/"+id, url.Values{}, &raw); err != nil {
		return Item{}, err
	}
	return mapShow(raw), nil
}

func (c *Client) GetEpisode(ctx context.Context, id string) (Item, error) {
	var raw episodeItem
	if err := c.get(ctx, "/episodes/"+id, url.Values{}, &raw); err != nil {
		return Item{}, err
	}
	return mapEpisode(raw), nil
}

func (c *Client) Playback(ctx context.Context) (PlaybackStatus, error) {
	var raw playbackResponse
	if err := c.get(ctx, "/me/player", nil, &raw); err != nil {
		if errors.Is(err, ErrNoContent) {
			return PlaybackStatus{IsPlaying: false}, nil
		}
		return PlaybackStatus{}, err
	}
	status := PlaybackStatus{
		IsPlaying:  raw.IsPlaying,
		ProgressMS: raw.ProgressMS,
		Shuffle:    raw.ShuffleState,
		Repeat:     raw.RepeatState,
		Device:     mapDevice(raw.Device),
	}
	if raw.Item.ID != "" {
		item := mapTrack(raw.Item)
		status.Item = &item
	}
	return status, nil
}

func (c *Client) Play(ctx context.Context, uri string) error {
	payload := map[string]any{}
	if uri != "" {
		if isContextURI(uri) {
			payload["context_uri"] = uri
		} else {
			payload["uris"] = []string{uri}
		}
	}
	return c.put(ctx, "/me/player/play", payload)
}

func (c *Client) Pause(ctx context.Context) error {
	return c.put(ctx, "/me/player/pause", nil)
}

func (c *Client) Next(ctx context.Context) error {
	return c.post(ctx, "/me/player/next", nil)
}

func (c *Client) Previous(ctx context.Context) error {
	return c.post(ctx, "/me/player/previous", nil)
}

func (c *Client) Seek(ctx context.Context, positionMS int) error {
	params := url.Values{}
	params.Set("position_ms", fmt.Sprint(positionMS))
	return c.putParams(ctx, "/me/player/seek", params)
}

func (c *Client) Volume(ctx context.Context, volume int) error {
	params := url.Values{}
	params.Set("volume_percent", fmt.Sprint(volume))
	return c.putParams(ctx, "/me/player/volume", params)
}

func (c *Client) Shuffle(ctx context.Context, enabled bool) error {
	params := url.Values{}
	params.Set("state", fmt.Sprint(enabled))
	return c.putParams(ctx, "/me/player/shuffle", params)
}

func (c *Client) Repeat(ctx context.Context, mode string) error {
	params := url.Values{}
	params.Set("state", mode)
	return c.putParams(ctx, "/me/player/repeat", params)
}

func (c *Client) Devices(ctx context.Context) ([]Device, error) {
	var raw deviceResponse
	if err := c.get(ctx, "/me/player/devices", nil, &raw); err != nil {
		return nil, err
	}
	devices := make([]Device, 0, len(raw.Devices))
	for _, d := range raw.Devices {
		devices = append(devices, mapDevice(d))
	}
	return devices, nil
}

func (c *Client) Transfer(ctx context.Context, deviceID string) error {
	payload := map[string]any{"device_ids": []string{deviceID}}
	return c.put(ctx, "/me/player", payload)
}

func (c *Client) QueueAdd(ctx context.Context, uri string) error {
	params := url.Values{}
	params.Set("uri", uri)
	return c.postParams(ctx, "/me/player/queue", params)
}

func (c *Client) Queue(ctx context.Context) (Queue, error) {
	var raw queueResponse
	if err := c.get(ctx, "/me/player/queue", nil, &raw); err != nil {
		if errors.Is(err, ErrNoContent) {
			return Queue{}, nil
		}
		return Queue{}, err
	}
	q := Queue{}
	if raw.CurrentlyPlaying.ID != "" {
		item := mapTrack(raw.CurrentlyPlaying)
		q.CurrentlyPlaying = &item
	}
	for _, item := range raw.Queue {
		q.Queue = append(q.Queue, mapTrack(item))
	}
	return q, nil
}

func (c *Client) LibraryTracks(ctx context.Context, limit, offset int) ([]Item, int, error) {
	return c.libraryTracks(ctx, "/me/tracks", limit, offset)
}

func (c *Client) LibraryAlbums(ctx context.Context, limit, offset int) ([]Item, int, error) {
	return c.libraryTracks(ctx, "/me/albums", limit, offset)
}

func (c *Client) libraryTracks(ctx context.Context, path string, limit, offset int) ([]Item, int, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprint(limit))
	params.Set("offset", fmt.Sprint(offset))
	var raw libraryResponse
	if err := c.get(ctx, path, params, &raw); err != nil {
		return nil, 0, err
	}
	items := make([]Item, 0, len(raw.Items))
	for _, item := range raw.Items {
		if item.Track.ID != "" {
			items = append(items, mapTrack(item.Track))
		}
		if item.Album.ID != "" {
			items = append(items, mapAlbum(item.Album))
		}
	}
	return items, raw.Total, nil
}

func (c *Client) LibraryModify(ctx context.Context, path string, ids []string, method string) error {
	params := url.Values{}
	params.Set("ids", joinComma(ids))
	return c.send(ctx, method, path, params, nil, nil)
}

func (c *Client) FollowArtists(ctx context.Context, ids []string, method string) error {
	params := url.Values{}
	params.Set("type", "artist")
	params.Set("ids", joinComma(ids))
	return c.send(ctx, method, "/me/following", params, nil, nil)
}

func (c *Client) Playlists(ctx context.Context, limit, offset int) ([]Item, int, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprint(limit))
	params.Set("offset", fmt.Sprint(offset))
	var raw playlistListResponse
	if err := c.get(ctx, "/me/playlists", params, &raw); err != nil {
		return nil, 0, err
	}
	items := make([]Item, 0, len(raw.Items))
	for _, item := range raw.Items {
		items = append(items, mapPlaylist(item))
	}
	return items, raw.Total, nil
}

func (c *Client) FollowedArtists(ctx context.Context, limit int, after string) ([]Item, int, string, error) {
	params := url.Values{}
	params.Set("type", "artist")
	params.Set("limit", fmt.Sprint(limit))
	if after != "" {
		params.Set("after", after)
	}
	var raw followedArtistsResponse
	if err := c.get(ctx, "/me/following", params, &raw); err != nil {
		return nil, 0, "", err
	}
	items := make([]Item, 0, len(raw.Artists.Items))
	for _, artist := range raw.Artists.Items {
		items = append(items, mapArtist(artist))
	}
	nextAfter := ""
	if len(raw.Artists.Items) > 0 {
		nextAfter = raw.Artists.Items[len(raw.Artists.Items)-1].ID
	}
	return items, raw.Artists.Total, nextAfter, nil
}

func (c *Client) PlaylistTracks(ctx context.Context, id string, limit, offset int) ([]Item, int, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprint(limit))
	params.Set("offset", fmt.Sprint(offset))
	var raw playlistTracksResponse
	if err := c.get(ctx, "/playlists/"+id+"/tracks", params, &raw); err != nil {
		return nil, 0, err
	}
	items := make([]Item, 0, len(raw.Items))
	for _, item := range raw.Items {
		if item.Track.ID == "" {
			continue
		}
		items = append(items, mapTrack(item.Track))
	}
	return items, raw.Total, nil
}

func (c *Client) CreatePlaylist(ctx context.Context, name string, public, collaborative bool) (Item, error) {
	userID, err := c.currentUserID(ctx)
	if err != nil {
		return Item{}, err
	}
	payload := map[string]any{
		"name":          name,
		"public":        public,
		"collaborative": collaborative,
	}
	var raw playlistItem
	if err := c.postJSON(ctx, "/users/"+userID+"/playlists", payload, &raw); err != nil {
		return Item{}, err
	}
	return mapPlaylist(raw), nil
}

func (c *Client) AddTracks(ctx context.Context, playlistID string, uris []string) error {
	payload := map[string]any{
		"uris": uris,
	}
	return c.postJSON(ctx, "/playlists/"+playlistID+"/tracks", payload, nil)
}

func (c *Client) RemoveTracks(ctx context.Context, playlistID string, uris []string) error {
	tracks := make([]map[string]string, 0, len(uris))
	for _, uri := range uris {
		tracks = append(tracks, map[string]string{"uri": uri})
	}
	payload := map[string]any{"tracks": tracks}
	return c.send(ctx, http.MethodDelete, "/playlists/"+playlistID+"/tracks", nil, payload, nil)
}

func (c *Client) currentUserID(ctx context.Context) (string, error) {
	var raw userProfile
	if err := c.get(ctx, "/me", nil, &raw); err != nil {
		return "", err
	}
	if raw.ID == "" {
		return "", errors.New("missing user id")
	}
	return raw.ID, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, dest any) error {
	return c.send(ctx, http.MethodGet, path, params, nil, dest)
}

func (c *Client) put(ctx context.Context, path string, payload any) error {
	return c.send(ctx, http.MethodPut, path, nil, payload, nil)
}

func (c *Client) post(ctx context.Context, path string, payload any) error {
	return c.send(ctx, http.MethodPost, path, nil, payload, nil)
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, dest any) error {
	return c.send(ctx, http.MethodPost, path, nil, payload, dest)
}

func (c *Client) putParams(ctx context.Context, path string, params url.Values) error {
	return c.send(ctx, http.MethodPut, path, params, nil, nil)
}

func (c *Client) postParams(ctx context.Context, path string, params url.Values) error {
	return c.send(ctx, http.MethodPost, path, params, nil, nil)
}

func (c *Client) send(ctx context.Context, method, path string, params url.Values, payload any, dest any) error {
	requestURL := c.baseURL + path
	if params == nil {
		if c.market != "" || c.language != "" || ((method == http.MethodPut || method == http.MethodPost || method == http.MethodDelete) && c.device != "") {
			params = url.Values{}
		}
	}
	if params != nil {
		if c.market != "" && params.Get("market") == "" {
			params.Set("market", c.market)
		}
		if c.language != "" && params.Get("locale") == "" {
			params.Set("locale", c.language)
		}
		if method == http.MethodPut || method == http.MethodPost || method == http.MethodDelete {
			if c.device != "" && params.Get("device_id") == "" {
				params.Set("device_id", c.device)
			}
		}
		if encoded := params.Encode(); encoded != "" {
			requestURL += "?" + encoded
		}
	}
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	token, err := c.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent())
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		if dest != nil {
			return ErrNoContent
		}
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiErrorFromResponse(resp)
	}
	if dest == nil {
		return nil
	}
	if resp.ContentLength == 0 {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *Client) token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastToken.AccessToken != "" && time.Until(c.lastToken.ExpiresAt) > time.Minute {
		return c.lastToken.AccessToken, nil
	}
	newToken, err := c.provider.Token(ctx)
	if err != nil {
		return "", err
	}
	c.lastToken = newToken
	return newToken.AccessToken, nil
}

func defaultUserAgent() string {
	return "spogo/0.1.0 (+https://github.com/steipete/spogo)"
}
