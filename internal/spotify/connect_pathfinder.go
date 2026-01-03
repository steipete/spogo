package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const pathfinderURL = "https://api-partner.spotify.com/pathfinder/v1/query"

func (c *ConnectClient) graphQL(ctx context.Context, operation string, variables map[string]any) (map[string]any, error) {
	auth, err := c.session.auth(ctx)
	if err != nil {
		return nil, err
	}
	hash, err := c.hashes.Hash(ctx, operation)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("operationName", operation)
	if variables == nil {
		variables = map[string]any{}
	}
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}
	extensionsJSON, err := json.Marshal(map[string]any{
		"persistedQuery": map[string]any{
			"version":    1,
			"sha256Hash": hash,
		},
	})
	if err != nil {
		return nil, err
	}
	params.Set("variables", string(variablesJSON))
	params.Set("extensions", string(extensionsJSON))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pathfinderURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("Client-Token", auth.ClientToken)
	req.Header.Set("Spotify-App-Version", auth.ClientVersion)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent())
	if c.language != "" {
		req.Header.Set("Accept-Language", c.language)
	}
	req.Header.Set("app-platform", "WebPlayer")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apiErrorFromResponse(resp)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if err := pathfinderError(payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func pathfinderError(payload map[string]any) error {
	errorsValue, ok := payload["errors"]
	if !ok {
		return nil
	}
	list, ok := errorsValue.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	first, ok := list[0].(map[string]any)
	if !ok {
		return errors.New("pathfinder error")
	}
	message, _ := first["message"].(string)
	if message == "" {
		message = "pathfinder error"
	}
	return errors.New(message)
}

func (c *ConnectClient) search(ctx context.Context, kind, query string, limit, offset int) (SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return SearchResult{}, errors.New("query required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	variables := map[string]any{
		"searchTerm": query,
		"offset":     offset,
		"limit":      limit,
	}
	payload, err := c.graphQL(ctx, "searchDesktop", variables)
	if err != nil {
		fallback, ferr := c.searchViaWeb(ctx, kind, query, limit, offset)
		if ferr == nil {
			return fallback, nil
		}
		return SearchResult{}, ferr
	}
	items, total := extractSearchItems(payload, kind)
	return SearchResult{
		Type:   kind,
		Limit:  limit,
		Offset: offset,
		Total:  total,
		Items:  items,
	}, nil
}

func (c *ConnectClient) trackInfo(ctx context.Context, id string) (Item, error) {
	item, err := c.infoByOperation(ctx, "getTrack", map[string]any{"uri": "spotify:track:" + id}, "track")
	if err == nil {
		return item, nil
	}
	web, ferr := c.webClient()
	if ferr != nil {
		return Item{}, err
	}
	return web.GetTrack(ctx, id)
}

func (c *ConnectClient) albumInfo(ctx context.Context, id string) (Item, error) {
	item, err := c.infoByOperation(ctx, "getAlbum", map[string]any{"uri": "spotify:album:" + id}, "album")
	if err == nil {
		return item, nil
	}
	web, ferr := c.webClient()
	if ferr != nil {
		return Item{}, err
	}
	return web.GetAlbum(ctx, id)
}

func (c *ConnectClient) artistInfo(ctx context.Context, id string) (Item, error) {
	item, err := c.infoByOperation(ctx, "queryArtistOverview", map[string]any{
		"uri":    "spotify:artist:" + id,
		"locale": c.language,
	}, "artist")
	if err == nil {
		return item, nil
	}
	web, ferr := c.webClient()
	if ferr != nil {
		return Item{}, err
	}
	return web.GetArtist(ctx, id)
}

func (c *ConnectClient) playlistInfo(ctx context.Context, id string) (Item, error) {
	item, err := c.infoByOperation(ctx, "fetchPlaylist", map[string]any{
		"uri":                       "spotify:playlist:" + id,
		"offset":                    0,
		"limit":                     25,
		"enableWatchFeedEntrypoint": false,
	}, "playlist")
	if err == nil {
		return item, nil
	}
	web, ferr := c.webClient()
	if ferr != nil {
		return Item{}, err
	}
	return web.GetPlaylist(ctx, id)
}

func (c *ConnectClient) showInfo(ctx context.Context, id string) (Item, error) {
	item, err := c.infoByOperation(ctx, "queryPodcastEpisodes", map[string]any{
		"uri":    "spotify:show:" + id,
		"offset": 0,
		"limit":  25,
	}, "show")
	if err == nil {
		return item, nil
	}
	web, ferr := c.webClient()
	if ferr != nil {
		return Item{}, err
	}
	return web.GetShow(ctx, id)
}

func (c *ConnectClient) episodeInfo(ctx context.Context, id string) (Item, error) {
	item, err := c.infoByOperation(ctx, "getEpisodeOrChapter", map[string]any{
		"uri": "spotify:episode:" + id,
	}, "episode")
	if err == nil {
		return item, nil
	}
	web, ferr := c.webClient()
	if ferr != nil {
		return Item{}, err
	}
	return web.GetEpisode(ctx, id)
}

func (c *ConnectClient) infoByOperation(ctx context.Context, operation string, variables map[string]any, kind string) (Item, error) {
	payload, err := c.graphQL(ctx, operation, variables)
	if err != nil {
		return Item{}, err
	}
	item, ok := extractItemFromPayload(payload, kind)
	if !ok {
		return Item{}, fmt.Errorf("no %s found", kind)
	}
	return item, nil
}

func (c *ConnectClient) searchViaWeb(ctx context.Context, kind, query string, limit, offset int) (SearchResult, error) {
	web, err := c.webClient()
	if err != nil {
		return SearchResult{}, err
	}
	return web.Search(ctx, kind, query, limit, offset)
}

func (c *ConnectClient) webClient() (*Client, error) {
	c.webMu.Lock()
	defer c.webMu.Unlock()
	if c.web != nil {
		return c.web, nil
	}
	provider := CookieTokenProvider{Source: c.source, Client: c.client}
	client, err := NewClient(Options{
		TokenProvider: provider,
		HTTPClient:    c.client,
		Market:        c.market,
		Language:      c.language,
		Device:        c.device,
	})
	if err != nil {
		return nil, err
	}
	c.web = client
	return client, nil
}
