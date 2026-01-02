package spotify

import "encoding/json"

func mapSearchItem(kind string, raw json.RawMessage) (Item, error) {
	switch kind {
	case "track":
		var t trackItem
		if err := json.Unmarshal(raw, &t); err != nil {
			return Item{}, err
		}
		return mapTrack(t), nil
	case "album":
		var a albumItem
		if err := json.Unmarshal(raw, &a); err != nil {
			return Item{}, err
		}
		return mapAlbum(a), nil
	case "artist":
		var a artistItem
		if err := json.Unmarshal(raw, &a); err != nil {
			return Item{}, err
		}
		return mapArtist(a), nil
	case "playlist":
		var p playlistItem
		if err := json.Unmarshal(raw, &p); err != nil {
			return Item{}, err
		}
		return mapPlaylist(p), nil
	case "show":
		var s showItem
		if err := json.Unmarshal(raw, &s); err != nil {
			return Item{}, err
		}
		return mapShow(s), nil
	case "episode":
		var e episodeItem
		if err := json.Unmarshal(raw, &e); err != nil {
			return Item{}, err
		}
		return mapEpisode(e), nil
	default:
		return Item{}, ErrUnsupportedType
	}
}

func mapTrack(t trackItem) Item {
	return Item{
		ID:         t.ID,
		URI:        t.URI,
		Name:       t.Name,
		Type:       "track",
		URL:        externalURL(t.ExternalURLs),
		Artists:    artistNames(t.Artists),
		Album:      t.Album.Name,
		DurationMS: t.DurationMS,
		Explicit:   t.Explicit,
		IsPlayable: t.IsPlayable,
	}
}

func mapAlbum(a albumItem) Item {
	return Item{
		ID:          a.ID,
		URI:         a.URI,
		Name:        a.Name,
		Type:        "album",
		URL:         externalURL(a.ExternalURLs),
		Artists:     artistNames(a.Artists),
		ReleaseDate: a.ReleaseDate,
		TotalTracks: a.TotalTracks,
	}
}

func mapArtist(a artistItem) Item {
	return Item{
		ID:        a.ID,
		URI:       a.URI,
		Name:      a.Name,
		Type:      "artist",
		URL:       externalURL(a.ExternalURLs),
		Followers: a.Followers.Total,
		Genres:    a.Genres,
	}
}

func mapPlaylist(p playlistItem) Item {
	return Item{
		ID:          p.ID,
		URI:         p.URI,
		Name:        p.Name,
		Type:        "playlist",
		URL:         externalURL(p.ExternalURLs),
		Owner:       p.Owner.DisplayName,
		TotalTracks: p.Tracks.Total,
		Description: p.Description,
	}
}

func mapShow(s showItem) Item {
	return Item{
		ID:            s.ID,
		URI:           s.URI,
		Name:          s.Name,
		Type:          "show",
		URL:           externalURL(s.ExternalURLs),
		Description:   s.Description,
		Publisher:     s.Publisher,
		TotalEpisodes: s.TotalEpisodes,
	}
}

func mapEpisode(e episodeItem) Item {
	return Item{
		ID:          e.ID,
		URI:         e.URI,
		Name:        e.Name,
		Type:        "episode",
		URL:         externalURL(e.ExternalURLs),
		Description: e.Description,
		DurationMS:  e.DurationMS,
	}
}

func mapDevice(d deviceItem) Device {
	return Device(d)
}

func artistNames(artists []artistRef) []string {
	ret := make([]string, 0, len(artists))
	for _, a := range artists {
		if a.Name == "" {
			continue
		}
		ret = append(ret, a.Name)
	}
	return ret
}

func externalURL(urls map[string]string) string {
	if urls == nil {
		return ""
	}
	if url := urls["spotify"]; url != "" {
		return url
	}
	for _, url := range urls {
		return url
	}
	return ""
}
