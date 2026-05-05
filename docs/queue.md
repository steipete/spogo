---
title: Queue
description: "Add tracks to and inspect the playback queue."
---

# Queue

The queue is the up-next list managed by Spotify Connect. It survives device transfers and persists across pause/resume.

## queue add

```bash
spogo queue add <id|url>
```

Appends one item to the queue. Accepts a track URI, URL, or bare ID (combine with `--type` for non-tracks):

```bash
spogo queue add spotify:track:7hQJA50XrCWABAu5v6QZ4i
spogo queue add https://open.spotify.com/track/4PTG3Z6ehGkBFwjybzWkR8
spogo queue add 0sf12qNH5qcw8qpgymFOqD --type track
```

`queue add` requires an active device. Open Spotify on a phone/desktop or pass `--device <name|id>`.

## queue show

```bash
spogo queue show
spogo queue show --plain
spogo queue show --json
```

Prints the currently-playing item plus the upcoming queue. Plain mode emits one item per line:

```
spotify:track:...   Track Name              Artist Name
spotify:track:...   Another Track           Another Artist
```

JSON mode includes `currently_playing` and a `queue` array with full track objects.

## queue clear

Spotify's API does not currently expose a way to clear the queue programmatically. The cleanest workaround is to start a new context, which replaces the queue:

```bash
spogo play spotify:track:7hQJA50XrCWABAu5v6QZ4i      # any single track
```

## Patterns

### Queue up the top results of a search

```bash
spogo search track "miles davis" --limit 5 --plain |
  awk '{print $1}' |
  while read uri; do spogo queue add "$uri"; done
```

### Queue an entire playlist's worth of next-up

`queue add` only takes one item — to queue every track from a playlist:

```bash
spogo playlist tracks "Road Trip" --plain |
  awk '{print $1}' |
  while read uri; do spogo queue add "$uri"; done
```

For long playlists this is N HTTP calls — usually faster to just `play` the playlist as a context.
