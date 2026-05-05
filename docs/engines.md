---
title: Engines
description: "Choose between connect, web, auto, and applescript — what each engine talks to and when to use it."
---

# Engines

spogo can talk to Spotify through one of four engines. Pick whichever matches what you need, or let `auto` decide.

## Quick pick

| Need | Engine |
| --- | --- |
| Default; works for almost everything | `connect` |
| Account where Connect is unavailable | `web` |
| "I just want it to work" | `auto` |
| Drive Spotify.app on macOS, no network needed | `applescript` |

Set with `--engine <name>` per call, or globally with `SPOGO_ENGINE`.

## connect (default)

Talks to Spotify's internal Connect endpoints — the same ones the official desktop and mobile apps use to coordinate playback across devices. spogo's first choice for everything.

**Best for**

- Playback control (play, pause, next, prev, seek, volume, shuffle, repeat).
- Device discovery and transfer.
- Playlist mutations under heavy use (Connect doesn't hit the Web API rate limits).
- Search and item info via the internal GraphQL surface.

**Tradeoffs**

- A handful of features fall back to the Web API automatically (e.g. transfers when Connect has no origin device, volume on certain hardware that needs `PUT`).
- Search/info uses GraphQL hashes; if a hash can't be resolved, falls back to web search.

## web

The public Spotify Web API. Slower, lower throughput, and rate-limited (~180 req/min/account before backoff), but it works on accounts where Connect is unavailable.

**Best for**

- Accounts that can't use Connect (rare — usually corporate or family-restricted).
- Forcing the documented public API for a reproducible test.
- Anything that requires Web API specific endpoints not yet in Connect.

**Tradeoffs**

- Rate limits will bite under bulk operations. If you see `429`, switch to `connect` or `auto`.
- Search/info/playback auto-fall-back to Connect when rate limited, so practical behavior is closer to `auto`.

## auto

Try `connect` first; fall back to `web` for unsupported features or when Connect signals the call won't work. The friendliest default if you're not sure.

```bash
spogo --engine auto play spotify:playlist:...
```

Most users don't need this — `connect` already falls back to web for the specific paths where it has to. `auto` is useful when you want **explicit** fallback behavior across all calls.

## applescript (macOS only)

Drives the local Spotify desktop app via AppleScript. No network, no cookies, no rate limits — but only the Mac you're on can be controlled, and you only see the local app's view (no Connect device list).

```bash
spogo --engine applescript play
spogo --engine applescript pause
spogo --engine applescript next
spogo --engine applescript status
```

**Best for**

- Quick local hotkeys / shortcuts (Raycast, Alfred, sketchybar, etc.) where network round trips are wasted.
- Sandboxed environments where cookie auth is awkward.
- Scripts that just need "pause my Mac's Spotify" without touching cloud state.

**Tradeoffs**

- macOS only.
- No Connect device list (`device list` shows just the Mac), no transfers.
- Search uses the Mac app's local search; results may differ from web search.
- Library/playlist mutations are not supported via AppleScript — fall back to `connect` or `web` for those.

## Setting an engine

Per command:

```bash
spogo --engine connect play
spogo --engine web search track "weezer"
spogo --engine applescript pause
```

Per shell:

```bash
export SPOGO_ENGINE=connect
```

In a config profile (`~/.config/spogo/<profile>/config.toml` or platform equivalent):

```toml
engine = "connect"
```

## Diagnosing engine issues

```bash
spogo --debug status
```

Debug logging on stderr shows which engine handled each call and any fallbacks that fired. See [Output](output.md) and [Troubleshooting](troubleshooting.md).
