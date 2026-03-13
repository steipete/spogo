package spotify

import "context"

func (c *ConnectClient) playlists(ctx context.Context, limit, offset int) ([]Item, int, error) {
	payload, err := c.graphQL(ctx, "libraryV3", libraryV3Variables("Playlists", normalizeLibraryLimit(limit), offset))
	if err != nil {
		return nil, 0, err
	}
	items, total := extractLibraryV3Items(payload, "playlist")
	return items, total, nil
}

func (c *ConnectClient) playlistTracks(ctx context.Context, id string, limit, offset int) ([]Item, int, error) {
	payload, err := c.graphQL(ctx, "fetchPlaylist", playlistTrackVariables(id, normalizePlaylistTrackLimit(limit), offset))
	if err != nil {
		return nil, 0, err
	}
	items, total := extractPlaylistContentItems(payload, "track")
	return items, total, nil
}

func (c *ConnectClient) libraryTracks(ctx context.Context, limit, offset int) ([]Item, int, error) {
	vars := map[string]any{
		"uri":    "spotify:collection:tracks",
		"offset": offset,
		"limit":  normalizeLibraryLimit(limit),
	}
	payload, err := c.graphQL(ctx, "fetchLibraryTracks", vars)
	if err != nil {
		return nil, 0, err
	}
	return extractFetchLibraryTracks(payload)
}

func (c *ConnectClient) libraryAlbums(ctx context.Context, limit, offset int) ([]Item, int, error) {
	payload, err := c.graphQL(ctx, "libraryV3", libraryV3Variables("Albums", normalizeLibraryLimit(limit), offset))
	if err != nil {
		return nil, 0, err
	}
	items, total := extractLibraryV3Items(payload, "album")
	return items, total, nil
}

func normalizeLibraryLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}

func normalizePlaylistTrackLimit(limit int) int {
	if limit <= 0 {
		return 25
	}
	return limit
}

func libraryV3Variables(filter string, limit, offset int) map[string]any {
	return map[string]any{
		"filters":                      []any{filter},
		"order":                        nil,
		"textFilter":                   "",
		"features":                     []any{"LIKED_SONGS", "YOUR_EPISODES"},
		"limit":                        limit,
		"offset":                       offset,
		"flatten":                      false,
		"expandedFolders":              []any{},
		"folderUri":                    nil,
		"includeFoldersWhenFlattening": true,
		"withCuration":                 false,
	}
}

func playlistTrackVariables(id string, limit, offset int) map[string]any {
	return map[string]any{
		"uri":                       "spotify:playlist:" + id,
		"offset":                    offset,
		"limit":                     limit,
		"enableWatchFeedEntrypoint": false,
	}
}
