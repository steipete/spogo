package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
	"github.com/steipete/spogo/internal/testutil"
)

func TestUserTopTracksCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			if timeRange != "long_term" {
				t.Fatalf("unexpected time_range: %s", timeRange)
			}
			return spotify.TopTracksResult{
				Total: 1, Limit: 20, Offset: 0,
				Items: []spotify.Item{{ID: "t1", Name: "Song", Type: "track", Artists: []string{"A"}, Album: "Album", URI: "spotify:track:t1"}},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "all-time", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestUserTopTracksCmdYear(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			if timeRange != "long_term" {
				t.Fatalf("expected long_term for year, got %s", timeRange)
			}
			return spotify.TopTracksResult{Total: 1, Limit: 20, Offset: 0, Items: []spotify.Item{{ID: "t1", Name: "Song", Type: "track"}}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "year", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestUserTopTracksCmd6mo(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			if timeRange != "medium_term" {
				t.Fatalf("expected medium_term, got %s", timeRange)
			}
			return spotify.TopTracksResult{Total: 1, Items: []spotify.Item{{ID: "t1", Name: "Song", Type: "track"}}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "6mo", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestUserTopTracksCmdMonth(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			if timeRange != "short_term" {
				t.Fatalf("expected short_term, got %s", timeRange)
			}
			return spotify.TopTracksResult{Total: 1, Items: []spotify.Item{{ID: "t1", Name: "Song", Type: "track"}}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "month", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestUserTopTracksCmdWeek(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			if timeRange != "short_term" {
				t.Fatalf("expected short_term, got %s", timeRange)
			}
			return spotify.TopTracksResult{Total: 1, Items: []spotify.Item{{ID: "t1", Name: "Song", Type: "track"}}}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "week", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestUserTopTracksCmdDayError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	cmd := UserTopTracksCmd{Period: "day", Limit: 20}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error for day period")
	}
}

func TestUserTopTracksCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			return spotify.TopTracksResult{}, errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "all-time", Limit: 20}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserTopTracksCmdJSON(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatJSON)
	mock := &testutil.SpotifyMock{
		GetUsersTopTracksFn: func(ctx context.Context, timeRange string, limit, offset int) (spotify.TopTracksResult, error) {
			return spotify.TopTracksResult{
				Total: 1, Limit: 20, Offset: 0,
				Items: []spotify.Item{{ID: "t1", Name: "Song", Type: "track", Artists: []string{"A"}, Album: "Album"}},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserTopTracksCmd{Period: "all-time", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), `"total": 1`) {
		t.Fatalf("JSON output missing total: %s", out.String())
	}
}

func TestUserHistoryCmd(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track", Artists: []string{"A"}, Album: "Album", URI: "spotify:track:t1"}, PlayedAt: "2024-01-15T10:00:00Z"},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestUserHistoryCmdWithPeriod(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			if after != 0 {
				t.Fatalf("period history should page with before and filter lower bound client-side, got after=%d", after)
			}
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track"}, PlayedAt: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "year", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestUserHistoryCmdCustomAfter(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	var capturedAfter int64
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			capturedAfter = after
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track"}, PlayedAt: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20, After: 1234567890000}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedAfter != 0 {
		t.Fatalf("expected API after=0; custom after is a client-side lower bound, got %d", capturedAfter)
	}
}

func TestUserHistoryCmdCustomBefore(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	var capturedBefore int64
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			capturedBefore = before
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track"}, PlayedAt: "2024-01-15T10:00:00Z"},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20, Before: 1234567890000}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedBefore != 1234567890000 {
		t.Fatalf("expected before=1234567890000, got %d", capturedBefore)
	}
}

func TestUserHistoryCmdError(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			return spotify.RecentlyPlayedResult{}, errors.New("boom")
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20}
	if err := cmd.Run(ctx); err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserHistoryCmdJSON(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatJSON)
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track", Artists: []string{"A"}}, PlayedAt: "2024-01-15T10:00:00Z"},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), `"played_at"`) {
		t.Fatalf("JSON output missing played_at: %s", out.String())
	}
}

func TestUserHistoryCmdPeriodAllNoAfter(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	var capturedAfter int64
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			capturedAfter = after
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track"}, PlayedAt: "2024-01-15T10:00:00Z"},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedAfter != 0 {
		t.Fatalf("expected after=0 for 'all' period, got %d", capturedAfter)
	}
}

func TestTopTrackPeriodMapping(t *testing.T) {
	tests := []struct {
		period   string
		expected string
	}{
		{"all-time", "long_term"},
		{"year", "long_term"},
		{"6mo", "medium_term"},
		{"month", "short_term"},
		{"week", "short_term"},
	}
	for _, tt := range tests {
		got := topTracksTimeRange(tt.period)
		if got != tt.expected {
			t.Errorf("topTracksTimeRange(%q) = %q, want %q", tt.period, got, tt.expected)
		}
	}
}

func TestHistoryPeriodProducesAfter(t *testing.T) {
	periods := []string{"year", "6mo", "1mo", "1wk", "1day"}
	for _, p := range periods {
		after := historyAfter(p)
		if after <= 0 {
			t.Errorf("historyAfter(%q) = %d, expected > 0", p, after)
		}
	}
}

func TestHistoryPeriodAllReturnsZero(t *testing.T) {
	if after := historyAfter("all"); after != 0 {
		t.Fatalf("historyAfter(all) = %d, expected 0", after)
	}
}

func TestUserHistoryCmdPagination(t *testing.T) {
	ctx, out, _ := testutil.NewTestContext(t, output.FormatPlain)
	playedAt1 := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	playedAt2 := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	callCount := 0
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			if after != 0 {
				t.Fatalf("pagination should use before cursor, got after=%d", after)
			}
			callCount++
			switch callCount {
			case 1:
				if before != 0 {
					t.Fatalf("expected first page to start from now, got before=%d", before)
				}
				return spotify.RecentlyPlayedResult{
					Limit: limit,
					Items: []spotify.RecentlyPlayedItem{
						{Track: spotify.Item{ID: "t1", Name: "Song1", Type: "track"}, PlayedAt: playedAt1},
					},
					Cursors: &spotify.Cursors{Before: "1705226400000"},
					Next:    "https://api.spotify.com/v1/me/player/recently-played?before=1705226400000",
				}, nil
			case 2:
				if before != 1705226400000 {
					t.Fatalf("expected second page before cursor, got %d", before)
				}
				return spotify.RecentlyPlayedResult{
					Limit: limit,
					Items: []spotify.RecentlyPlayedItem{
						{Track: spotify.Item{ID: "t2", Name: "Song2", Type: "track"}, PlayedAt: playedAt2},
					},
					Cursors: &spotify.Cursors{Before: "1705140000000"},
					Next:    "https://api.spotify.com/v1/me/player/recently-played?before=1705140000000",
				}, nil
			default:
				return spotify.RecentlyPlayedResult{Limit: limit}, nil
			}
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "year", Limit: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if callCount != 3 {
		t.Fatalf("expected 3 API calls for pagination, got %d", callCount)
	}
	if out.String() == "" {
		t.Fatalf("expected output")
	}
}

func TestUserHistoryCmdPaginationStopsAtCap(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	playedAt := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	callCount := 0
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			callCount++
			items := make([]spotify.RecentlyPlayedItem, limit)
			for i := range items {
				items[i] = spotify.RecentlyPlayedItem{
					Track:    spotify.Item{ID: fmt.Sprintf("t%d", callCount*limit+i), Name: "Song", Type: "track"},
					PlayedAt: playedAt,
				}
			}
			return spotify.RecentlyPlayedResult{
				Limit:   limit,
				Items:   items,
				Cursors: &spotify.Cursors{Before: fmt.Sprintf("%d", 1700000000000-callCount)},
				Next:    "https://api.spotify.com/v1/me/player/recently-played?before=next",
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "year", Limit: 50}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if callCount != 4 {
		t.Fatalf("expected 4 API calls (50*4=200 cap), got %d", callCount)
	}
}

func TestUserHistoryCmdBeforeOnly(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	var capturedBefore int64
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			capturedBefore = before
			if after != 0 {
				t.Fatalf("expected after=0 for before-only request, got %d", after)
			}
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "Song", Type: "track"}, PlayedAt: "2024-01-15T10:00:00Z"},
				},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "all", Limit: 20, Before: 1234567890000}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedBefore != 1234567890000 {
		t.Fatalf("expected before=1234567890000, got %d", capturedBefore)
	}
}

func TestUserHistoryCmdAfterAndBeforeFilter(t *testing.T) {
	ctx, _, _ := testutil.NewTestContext(t, output.FormatPlain)
	callCount := 0
	mock := &testutil.SpotifyMock{
		GetRecentlyPlayedFn: func(ctx context.Context, limit int, after, before int64) (spotify.RecentlyPlayedResult, error) {
			if after != 0 {
				t.Fatalf("expected after=0; pagination should use before and filter lower bound client-side, got %d", after)
			}
			callCount++
			if callCount == 1 && before != 1705363200000 {
				t.Fatalf("expected first request to honor explicit before, got %d", before)
			}
			return spotify.RecentlyPlayedResult{
				Limit: limit,
				Items: []spotify.RecentlyPlayedItem{
					{Track: spotify.Item{ID: "t1", Name: "S1", Type: "track"}, PlayedAt: "2024-01-15T10:00:00Z"},
					{Track: spotify.Item{ID: "t2", Name: "S2", Type: "track"}, PlayedAt: "2024-01-16T10:00:00Z"},
				},
				Cursors: &spotify.Cursors{Before: "1705312800000"},
			}, nil
		},
	}
	ctx.SetSpotify(mock)
	cmd := UserHistoryCmd{Period: "year", Limit: 20, Before: 1705363200000}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestParseRFC3339Milli(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"2024-01-15T10:00:00Z", 1705312800000},
		{"2024-01-15T10:00:00.000Z", 1705312800000},
	}
	for _, tt := range tests {
		got, err := parseRFC3339Milli(tt.input)
		if err != nil {
			t.Fatalf("parseRFC3339Milli(%q): %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("parseRFC3339Milli(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseRFC3339MilliError(t *testing.T) {
	if _, err := parseRFC3339Milli("invalid"); err == nil {
		t.Fatalf("expected error for invalid input")
	}
}
