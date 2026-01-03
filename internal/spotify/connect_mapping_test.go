package spotify

import "testing"

func TestExtractItem(t *testing.T) {
	raw := map[string]any{
		"uri":  "spotify:track:abc",
		"name": "Song",
		"artists": []any{
			map[string]any{"name": "Artist"},
		},
		"album": map[string]any{"name": "Album"},
	}
	item, ok := extractItem(raw, "track")
	if !ok {
		t.Fatalf("expected item")
	}
	if item.ID != "abc" || item.Name != "Song" || item.Album != "Album" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if len(item.Artists) != 1 || item.Artists[0] != "Artist" {
		t.Fatalf("unexpected artists: %#v", item.Artists)
	}
}

func TestSearchPaths(t *testing.T) {
	paths := searchPaths("track")
	if len(paths) == 0 {
		t.Fatalf("expected paths")
	}
}

func TestExtractItemFallbacks(t *testing.T) {
	raw := map[string]any{
		"id":    "t1",
		"title": "Song",
		"artists": []any{
			map[string]any{"name": "Artist"},
			map[string]any{"name": "Artist"},
		},
		"albumOfTrack": map[string]any{"name": "Album"},
		"owner":        map[string]any{"name": "Owner"},
		"durationMs":   1200,
		"explicit":     true,
		"total":        5,
		"releaseDate":  "2020-01-01",
		"description":  "desc",
		"isPlayable":   true,
	}
	item, ok := extractItem(raw, "track")
	if !ok {
		t.Fatalf("expected item")
	}
	if item.URI != "spotify:track:t1" || item.Name != "Song" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if len(item.Artists) != 1 || item.Artists[0] != "Artist" {
		t.Fatalf("expected deduped artists: %#v", item.Artists)
	}
	if item.Album != "Album" || item.Owner != "Owner" || item.DurationMS != 1200 {
		t.Fatalf("unexpected fields: %#v", item)
	}
}

func TestExtractSearchItemsFallback(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"nested": map[string]any{
				"item": map[string]any{"uri": "spotify:track:abc", "name": "Song"},
			},
		},
	}
	items, total := extractSearchItems(payload, "track")
	if len(items) == 0 || items[0].URI != "spotify:track:abc" {
		t.Fatalf("unexpected items: %#v", items)
	}
	if total != len(items) {
		t.Fatalf("expected total to match items")
	}
}

func TestSearchPathsVariants(t *testing.T) {
	if len(searchPaths("album")) == 0 {
		t.Fatalf("expected album paths")
	}
	if len(searchPaths("artist")) == 0 {
		t.Fatalf("expected artist paths")
	}
	if len(searchPaths("playlist")) == 0 {
		t.Fatalf("expected playlist paths")
	}
	if len(searchPaths("show")) == 0 {
		t.Fatalf("expected show paths")
	}
	if len(searchPaths("episode")) == 0 {
		t.Fatalf("expected episode paths")
	}
	if searchPaths("unknown") != nil {
		t.Fatalf("expected nil paths")
	}
}

func TestHelperFinds(t *testing.T) {
	value := map[string]any{
		"nested": []any{
			map[string]any{"title": "Hello"},
			map[string]any{"uri": "spotify:track:abc"},
		},
	}
	if name := findFirstName(value); name != "Hello" {
		t.Fatalf("unexpected name: %s", name)
	}
	if uri := findFirstURI(value, "track"); uri != "spotify:track:abc" {
		t.Fatalf("unexpected uri: %s", uri)
	}
	payload := map[string]any{"a": map[string]any{"b": map[string]any{"value": "ok", "num": 2.0, "flag": true}}}
	if m, ok := getMap(payload, "a", "b"); !ok || getString(m, "value") != "ok" {
		t.Fatalf("unexpected map")
	}
	m, _ := getMap(payload, "a", "b")
	if getInt(m, "num") != 2 || !getBool(m, "flag") {
		t.Fatalf("unexpected helpers")
	}
	if idFromURI("abc") != "abc" {
		t.Fatalf("unexpected id")
	}
	if typeFromURI("abc") != "" {
		t.Fatalf("unexpected type")
	}
}

func TestVisitItemsSlice(t *testing.T) {
	values := []any{
		map[string]any{"uri": "spotify:track:abc", "name": "Song"},
	}
	items := []Item{}
	visitItems(values, "track", &items)
	if len(items) != 1 || items[0].ID != "abc" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestHelperNilAndOwnerFallback(t *testing.T) {
	if getString(nil, "key") != "" {
		t.Fatalf("expected empty string")
	}
	if getInt(nil, "key") != 0 {
		t.Fatalf("expected zero int")
	}
	if getBool(nil, "key") {
		t.Fatalf("expected false bool")
	}
	value := map[string]any{
		"user": map[string]any{"name": "Owner"},
	}
	if name := extractOwnerName(value); name != "Owner" {
		t.Fatalf("unexpected owner: %s", name)
	}
	container := map[string]any{
		"item": map[string]any{"uri": "spotify:track:q1", "name": "Song"},
	}
	items := extractItemsFromContainer(container, "track")
	if len(items) == 0 || items[0].ID != "q1" {
		t.Fatalf("unexpected items: %#v", items)
	}
}
