package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
)

const maxHistoryItems = 200

type UserCmd struct {
	TopTracks UserTopTracksCmd `kong:"cmd,help='Show your top tracks by affinity ranking.'"`
	History   UserHistoryCmd   `kong:"cmd,help='Show recently played tracks available from Spotify.'"`
}

type UserTopTracksCmd struct {
	Period string `help:"Affinity period: all-time/year use Spotify long_term, 6mo uses medium_term, month/week use short_term (~4 weeks), day is unsupported." default:"all-time" enum:"all-time,year,6mo,month,week,day"`
	Limit  int    `help:"Number of results." default:"20"`
	Offset int    `help:"Offset results." default:"0"`
}

type UserHistoryCmd struct {
	Period string `help:"Local lower-bound filter over Spotify's available recent plays: all, year, 6mo, 1mo, 1wk, 1day." default:"all" enum:"all,year,6mo,1mo,1wk,1day"`
	Limit  int    `help:"Items per page (max 50)." default:"20"`
	After  int64  `help:"Client-side lower-bound filter, Unix timestamp (ms). Pagination still walks backward with before cursors."`
	Before int64  `help:"Return items played before this Unix timestamp (ms)."`
}

var topTrackPeriods = map[string]string{
	"all-time": "long_term",
	"year":     "long_term",
	"6mo":      "medium_term",
	"month":    "short_term",
	"week":     "short_term",
}

func topTracksTimeRange(period string) string {
	return topTrackPeriods[period]
}

func topTracksPeriodNote(period string) string {
	switch period {
	case "year":
		return "Spotify has no calendar-year top-tracks window; year maps to long_term (lifetime affinity)."
	case "month":
		return "Spotify short_term is roughly the last 4 weeks; this is an approximation."
	case "week":
		return "Spotify has no weekly top-tracks window; week maps to short_term (roughly last 4 weeks)."
	default:
		return "Spotify returns affinity-ranked top tracks, not play counts."
	}
}

var historyPeriodDurations = map[string]time.Duration{
	"year": 365 * 24 * time.Hour,
	"6mo":  180 * 24 * time.Hour,
	"1mo":  30 * 24 * time.Hour,
	"1wk":  7 * 24 * time.Hour,
	"1day": 24 * time.Hour,
}

func historyAfter(period string) int64 {
	d, ok := historyPeriodDurations[period]
	if !ok {
		return 0
	}
	return time.Now().Add(-d).UnixMilli()
}

func (cmd *UserTopTracksCmd) Run(ctx *app.Context) error {
	if cmd.Period == "day" {
		return fmt.Errorf("Spotify does not support daily top tracks; only long_term, medium_term, short_term; use all-time, year, 6mo, month, or week (week maps to short_term, ~4 weeks)")
	}
	client, cmdCtx, err := spotifyClient(ctx)
	if err != nil {
		return err
	}
	timeRange := topTracksTimeRange(cmd.Period)
	limit := clampLimit(cmd.Limit)
	res, err := client.GetUsersTopTracks(cmdCtx, timeRange, limit, cmd.Offset)
	if err != nil {
		return err
	}
	plain, human := renderTopTracks(ctx.Output, res.Items)
	payload := map[string]any{
		"total":      res.Total,
		"limit":      res.Limit,
		"offset":     res.Offset,
		"items":      res.Items,
		"time_range": timeRange,
		"period":     cmd.Period,
		"note":       topTracksPeriodNote(cmd.Period),
		"period_map": topTrackPeriods,
	}
	if ctx.Output.Format == output.FormatHuman {
		header := fmt.Sprintf("Top tracks (%s -> %s): %d (affinity ranking, not play counts)", cmd.Period, timeRange, res.Total)
		human = append([]string{header, topTracksPeriodNote(cmd.Period)}, human...)
	}
	return ctx.Output.Emit(payload, plain, human)
}

func (cmd *UserHistoryCmd) Run(ctx *app.Context) error {
	client, cmdCtx, err := spotifyClient(ctx)
	if err != nil {
		return err
	}

	limit := clampLimit(cmd.Limit)
	after := cmd.After
	if after == 0 && cmd.Period != "all" {
		after = historyAfter(cmd.Period)
	}

	var allItems []spotify.RecentlyPlayedItem
	var lastCursors *spotify.Cursors
	before := cmd.Before
	stopAtLowerBound := false

	for {
		res, err := client.GetRecentlyPlayed(cmdCtx, limit, 0, before)
		if err != nil {
			return err
		}
		if res.Cursors != nil {
			lastCursors = res.Cursors
		}
		for _, item := range res.Items {
			ms, err := parseRFC3339Milli(item.PlayedAt)
			if err != nil {
				continue
			}
			if cmd.Before > 0 && ms >= cmd.Before {
				continue
			}
			if after > 0 && ms < after {
				stopAtLowerBound = true
				continue
			}
			allItems = append(allItems, item)
			if len(allItems) >= maxHistoryItems {
				break
			}
		}
		if len(res.Items) == 0 || res.Cursors == nil || res.Cursors.Before == "" || len(allItems) >= maxHistoryItems || stopAtLowerBound {
			break
		}
		nextBefore, err := strconv.ParseInt(res.Cursors.Before, 10, 64)
		if err != nil || nextBefore == before {
			break
		}
		before = nextBefore
	}

	plain, human := renderRecentlyPlayed(ctx.Output, allItems)
	payload := map[string]any{
		"items":                 allItems,
		"cursors":               lastCursors,
		"total_fetched":         len(allItems),
		"max_allowed":           maxHistoryItems,
		"period":                cmd.Period,
		"requested_after":       cmd.After,
		"requested_before":      cmd.Before,
		"effective_lower_bound": after,
		"note":                  "Spotify recently-played returns available recent plays only; periods are client-side filters, not complete historical archives.",
	}
	if ctx.Output.Format == output.FormatHuman {
		header := fmt.Sprintf("Recently played available from Spotify: %d", len(allItems))
		if cmd.Period != "all" {
			header = fmt.Sprintf("Recently played available within %s: %d", cmd.Period, len(allItems))
		}
		human = append([]string{header, "Spotify retention and the 200-item client cap mean this is not a complete listening archive."}, human...)
	}
	return ctx.Output.Emit(payload, plain, human)
}

func parseRFC3339Milli(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
		if err != nil {
			return 0, err
		}
	}
	return t.UnixMilli(), nil
}

func renderTopTracks(w *output.Writer, items []spotify.Item) (plain []string, human []string) {
	accent := w.Theme.Accent
	muted := w.Theme.Muted
	plain = make([]string, 0, len(items))
	human = make([]string, 0, len(items))
	for i, item := range items {
		rank := i + 1
		plain = append(plain, fmt.Sprintf("%d\ttrack\t%s\t%s\t%s\t%s\t%s", rank, item.ID, item.Name, strings.Join(item.Artists, ", "), item.Album, item.URI))
		human = append(human, fmt.Sprintf("%d. %s — %s %s", rank, accent(item.Name), strings.Join(item.Artists, ", "), muted("· "+item.Album)))
	}
	return plain, human
}

func renderRecentlyPlayed(w *output.Writer, items []spotify.RecentlyPlayedItem) (plain []string, human []string) {
	accent := w.Theme.Accent
	muted := w.Theme.Muted
	plain = make([]string, 0, len(items))
	human = make([]string, 0, len(items))
	for _, item := range items {
		plain = append(plain, fmt.Sprintf("%s\ttrack\t%s\t%s\t%s\t%s\t%s", item.PlayedAt, item.Track.ID, item.Track.Name, strings.Join(item.Track.Artists, ", "), item.Track.Album, item.Track.URI))
		human = append(human, fmt.Sprintf("%s — %s %s", accent(item.Track.Name), strings.Join(item.Track.Artists, ", "), muted("· "+item.PlayedAt)))
	}
	return plain, human
}
