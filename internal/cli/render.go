package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
)

func renderItems(w *output.Writer, items []spotify.Item) (plain []string, human []string) {
	plain = make([]string, 0, len(items))
	human = make([]string, 0, len(items))
	for _, item := range items {
		plain = append(plain, itemPlain(item))
		human = append(human, itemHuman(w, item))
	}
	return plain, human
}

func itemPlain(item spotify.Item) string {
	switch item.Type {
	case "track":
		return fmt.Sprintf("track\t%s\t%s\t%s\t%s\t%s", item.ID, item.Name, strings.Join(item.Artists, ", "), item.Album, item.URI)
	case "album":
		return fmt.Sprintf("album\t%s\t%s\t%s\t%s\t%d", item.ID, item.Name, strings.Join(item.Artists, ", "), item.ReleaseDate, item.TotalTracks)
	case "artist":
		return fmt.Sprintf("artist\t%s\t%s\t%d", item.ID, item.Name, item.Followers)
	case "playlist":
		return fmt.Sprintf("playlist\t%s\t%s\t%s\t%d", item.ID, item.Name, item.Owner, item.TotalTracks)
	case "show":
		return fmt.Sprintf("show\t%s\t%s\t%s\t%d", item.ID, item.Name, item.Publisher, item.TotalEpisodes)
	case "episode":
		return fmt.Sprintf("episode\t%s\t%s\t%d", item.ID, item.Name, item.DurationMS)
	default:
		return fmt.Sprintf("item\t%s\t%s\t%s", item.ID, item.Name, item.URI)
	}
}

func itemHuman(w *output.Writer, item spotify.Item) string {
	accent := w.Theme.Accent
	muted := w.Theme.Muted
	switch item.Type {
	case "track":
		return fmt.Sprintf("%s — %s %s", accent(item.Name), strings.Join(item.Artists, ", "), muted("· "+item.Album))
	case "album":
		return fmt.Sprintf("%s — %s %s", accent(item.Name), strings.Join(item.Artists, ", "), muted("· "+item.ReleaseDate))
	case "artist":
		return fmt.Sprintf("%s %s", accent(item.Name), muted(fmt.Sprintf("· %d followers", item.Followers)))
	case "playlist":
		return fmt.Sprintf("%s — %s %s", accent(item.Name), item.Owner, muted(fmt.Sprintf("· %d tracks", item.TotalTracks)))
	case "show":
		return fmt.Sprintf("%s — %s %s", accent(item.Name), item.Publisher, muted(fmt.Sprintf("· %d episodes", item.TotalEpisodes)))
	case "episode":
		return fmt.Sprintf("%s %s", accent(item.Name), muted(fmt.Sprintf("· %s", humanDuration(item.DurationMS))))
	default:
		return accent(item.Name)
	}
}

func humanDuration(ms int) string {
	if ms <= 0 {
		return "0s"
	}
	d := time.Duration(ms) * time.Millisecond
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func playbackPlain(status spotify.PlaybackStatus) string {
	track := ""
	if status.Item != nil {
		track = status.Item.Name
	}
	return fmt.Sprintf("%t\t%d\t%s\t%s", status.IsPlaying, status.ProgressMS, status.Device.Name, track)
}

func playbackHuman(w *output.Writer, status spotify.PlaybackStatus) string {
	accent := w.Theme.Accent
	muted := w.Theme.Muted
	state := "paused"
	if status.IsPlaying {
		state = "playing"
	}
	track := ""
	if status.Item != nil {
		track = fmt.Sprintf("%s — %s", accent(status.Item.Name), strings.Join(status.Item.Artists, ", "))
	}
	return fmt.Sprintf("%s %s %s", accent(strings.ToUpper(state)), track, muted("· "+status.Device.Name))
}
