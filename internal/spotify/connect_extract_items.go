package spotify

import (
	"fmt"
	"strings"
)

func extractSearchItems(payload map[string]any, kind string) ([]Item, int) {
	paths := searchPaths(kind)
	for _, path := range paths {
		if container, ok := getMap(payload, path...); ok {
			items := extractItemsFromContainer(container, kind)
			total := getInt(container, "totalCount")
			if total == 0 {
				total = len(items)
			}
			return items, total
		}
	}
	items := collectItemsByKind(payload, kind)
	return items, len(items)
}

func extractItemFromPayload(payload map[string]any, kind string) (Item, bool) {
	if kind == "track" {
		if m, ok := getMap(payload, "data", "trackUnion"); ok {
			if item, ok := extractItem(m, kind); ok {
				return item, true
			}
		}
		if m, ok := getMap(payload, "data", "track"); ok {
			if item, ok := extractItem(m, kind); ok {
				return item, true
			}
		}
	}
	items := collectItemsByKind(payload, kind)
	if len(items) == 0 {
		return Item{}, false
	}
	return items[0], true
}

func searchPaths(kind string) [][]string {
	switch kind {
	case "track":
		return [][]string{{"data", "searchV2", "tracksV2"}}
	case "album":
		return [][]string{
			{"data", "searchV2", "albumsV2"},
			{"data", "searchV2", "albums"},
		}
	case "artist":
		return [][]string{{"data", "searchV2", "artists"}}
	case "playlist":
		return [][]string{{"data", "searchV2", "playlists"}}
	case "show":
		return [][]string{
			{"data", "searchV2", "podcasts"},
			{"data", "searchV2", "shows"},
		}
	case "episode":
		return [][]string{{"data", "searchV2", "episodes"}}
	default:
		return nil
	}
}

func extractItemsFromContainer(container map[string]any, kind string) []Item {
	itemsRaw, ok := container["items"].([]any)
	if !ok {
		return collectItemsByKind(container, kind)
	}
	items := make([]Item, 0, len(itemsRaw))
	for _, raw := range itemsRaw {
		item, ok := extractItem(raw, kind)
		if ok {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return collectItemsByKind(container, kind)
	}
	return items
}

func collectItemsByKind(value any, kind string) []Item {
	items := []Item{}
	visitItems(value, kind, &items)
	return items
}

func visitItems(value any, kind string, items *[]Item) {
	switch typed := value.(type) {
	case map[string]any:
		if item, ok := extractItem(typed, kind); ok {
			*items = append(*items, item)
		}
		for _, child := range typed {
			visitItems(child, kind, items)
		}
	case []any:
		for _, child := range typed {
			visitItems(child, kind, items)
		}
	}
}

func extractItem(value any, kind string) (Item, bool) {
	m, ok := value.(map[string]any)
	if !ok {
		return Item{}, false
	}
	if kind == "track" {
		if inner, ok := m["track"].(map[string]any); ok {
			m = inner
		}
	}
	uri := getString(m, "uri")
	if uri == "" && kind != "" {
		if id := getString(m, "id"); id != "" {
			uri = "spotify:" + kind + ":" + id
		}
	}
	if uri == "" {
		if inner := findFirstURI(m, kind); inner != "" {
			uri = inner
		}
	}
	if uri == "" {
		return Item{}, false
	}
	if kind != "" && !strings.HasPrefix(uri, "spotify:"+kind+":") {
		return Item{}, false
	}
	name := getString(m, "name")
	if name == "" {
		name = getString(m, "title")
	}
	if name == "" {
		name = findFirstName(m)
	}
	item := Item{
		URI:  uri,
		ID:   idFromURI(uri),
		Name: name,
		Type: typeFromURI(uri),
	}
	item.URL = fmt.Sprintf("https://open.spotify.com/%s/%s", item.Type, item.ID)
	item.Artists = extractArtistNames(m)
	if album := extractAlbumName(m); album != "" {
		item.Album = album
	}
	item.Explicit = getBool(m, "explicit")
	item.DurationMS = getInt(m, "duration_ms")
	if item.DurationMS == 0 {
		item.DurationMS = getInt(m, "durationMs")
	}
	item.Owner = extractOwnerName(m)
	item.TotalTracks = getInt(m, "totalTracks")
	if item.TotalTracks == 0 {
		item.TotalTracks = getInt(m, "total")
	}
	item.ReleaseDate = getString(m, "releaseDate")
	item.Description = getString(m, "description")
	item.IsPlayable = getBool(m, "isPlayable")
	item.Publisher = getString(m, "publisher")
	item.TotalEpisodes = getInt(m, "totalEpisodes")
	return item, true
}

func idFromURI(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return uri
}

func typeFromURI(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) >= 3 {
		return parts[len(parts)-2]
	}
	return ""
}

func findFirstURI(value any, kind string) string {
	switch typed := value.(type) {
	case map[string]any:
		if uri, ok := typed["uri"].(string); ok {
			if kind == "" || strings.HasPrefix(uri, "spotify:"+kind+":") {
				return uri
			}
		}
		for _, child := range typed {
			if uri := findFirstURI(child, kind); uri != "" {
				return uri
			}
		}
	case []any:
		for _, child := range typed {
			if uri := findFirstURI(child, kind); uri != "" {
				return uri
			}
		}
	}
	return ""
}

func findFirstName(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		if name, ok := typed["name"].(string); ok {
			return name
		}
		if title, ok := typed["title"].(string); ok {
			return title
		}
		for _, child := range typed {
			if name := findFirstName(child); name != "" {
				return name
			}
		}
	case []any:
		for _, child := range typed {
			if name := findFirstName(child); name != "" {
				return name
			}
		}
	}
	return ""
}
