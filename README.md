# 🎧 spogo - Spotify, but make it terminal.

 Power CLI using web cookies. Search, control playback, manage library/playlists, and script with JSON/plain output.

## Why Cookies?

Spotify's official API has strict rate limits that make it impractical for agents and automation. By using browser cookies to authenticate with Spotify's internal web API (the same one their web player uses), spogo bypasses these limitations:

- **No rate limits** - Use the same endpoints as open.spotify.com
- **No app registration** - No need to create a Spotify Developer app
- **Full functionality** - Access to everything the web player can do
- **Agent-friendly** - Perfect for AI assistants and automation scripts

Import your cookies once with `sweetcookie` and you're good to go.

## Features

- Search tracks, albums, artists, playlists, shows, episodes
- Playback control: play/pause/next/prev/seek/volume/shuffle/repeat
- Queue management
- Library management (save/remove/follow)
- Playlist management (create/add/remove/list)
- Device selection and status
- Browser cookie import via `sweetcookie` (local `sweetcookie`)
- `--json` and `--plain` for scripting
- Colorized human output (respects `NO_COLOR`, `TERM=dumb`, `--no-color`)

## Install

```bash
go install github.com/steipete/spogo/cmd/spogo@latest
```

## Quick start

```bash
spogo auth import --browser chrome
spogo auth import --browser chrome --browser-profile "Profile 1"
spogo search track "weezer" --limit 5
spogo play spotify:track:7hQJA50XrCWABAu5v6QZ4i
spogo status
```

## Usage

```bash
spogo [global flags] <command> [args]
```

Global flags:

- `--config <path>` config file path
- `--profile <name>` profile name
- `--timeout <dur>` request timeout (default 10s)
- `--market <cc>` market country code
- `--language <tag>` language/locale (default `en`)
- `--device <name|id>` target device
- `--json` / `--plain`
- `--no-color`
- `-q, --quiet` / `-v, --verbose` / `-d, --debug`

Commands:

- `auth status|import|clear`
- `search track|album|artist|playlist|show|episode`
- `track info`, `album info`, `artist info`, `playlist info`, `show info`, `episode info`
- `play [<id|url>] [--type ...]`, `pause`, `next`, `prev`, `seek`, `volume`, `shuffle`, `repeat`, `status`
- `queue add|show`
- `library tracks|albums|artists|playlists`
- `playlist create|add|remove|tracks`
- `device list|set`

Full spec: `docs/spec.md`.

## Cookies

`spogo` uses browser cookies (via `sweetcookie`) to fetch a web access token. Import cookies once:

```bash
spogo auth import --browser chrome
```

Cookies are stored under your config directory (per profile).

## Output

- Human output by default
- `--plain` for line-oriented output
- `--json` for structured output

## Legal

This tool interacts with Spotify’s web endpoints. Use responsibly and in accordance with Spotify’s Terms of Service.

## License

MIT
