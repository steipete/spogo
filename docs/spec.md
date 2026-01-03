# spogo CLI spec (v0.1.0)

One-liner: Spotify power CLI using web cookies; search + playback control.
Parser: Kong.
Cookies: steipete/sweetcookie (local sweetcookie).
Output: human by default; `--plain` or `--json`.
Color: on by default; respects `NO_COLOR`, `TERM=dumb`, `--no-color`.
Platforms: macOS, Linux, Windows.

## Usage

```
spogo [global flags] <command> [args]
```

## Global flags

- `-h, --help`
- `--version`
- `-q, --quiet`
- `-v, --verbose`
- `-d, --debug`
- `--json`
- `--plain`
- `--no-color`
- `--config <path>` default: `os.UserConfigDir()/spogo/config.toml`
- `--profile <name>` default: `default`
- `--timeout <dur>` default: `10s`
- `--market <cc>` default: account market or `US`
- `--language <tag>` default: `en`
- `--device <name|id>` default: active device
- `--engine <web|connect>` default: `connect`
- `--no-input`

## Commands

### auth

- `spogo auth status`
- `spogo auth import`
  - flags: `--browser <chrome|brave|edge|firefox|safari>` default: `chrome`
  - `--browser-profile <name>`
  - `--cookie-path <file>`
  - `--domain <host>` default `spotify.com`
- `spogo auth clear`

### search

- `spogo search track <query> [--limit N] [--offset N]`
- `spogo search album <query> [--limit N] [--offset N]`
- `spogo search artist <query> [--limit N] [--offset N]`
- `spogo search playlist <query> [--limit N] [--offset N]`
- `spogo search episode <query> [--limit N] [--offset N]`
- `spogo search show <query> [--limit N] [--offset N]`

### info

- `spogo track info <id|url>`
- `spogo album info <id|url>`
- `spogo artist info <id|url>`
- `spogo playlist info <id|url>`
- `spogo show info <id|url>`
- `spogo episode info <id|url>`

### playback

- `spogo play [<id|url>]` (track/album/playlist/show)
  - optional: `--type <track|album|playlist|show|episode>` for raw IDs
- `spogo pause`
- `spogo next`
- `spogo prev`
- `spogo seek <ms|mm:ss>`
- `spogo volume <0-100>`
- `spogo shuffle <on|off>`
- `spogo repeat <off|track|context>`
- `spogo status`

### queue

- `spogo queue add <id|url>`
- `spogo queue show`
- `spogo queue clear` (not supported by Spotify API yet)

### library

- `spogo library tracks list [--limit N]`
- `spogo library tracks add <id|url...>`
- `spogo library tracks remove <id|url...>`
- `spogo library albums list [--limit N]`
- `spogo library albums add <id|url...>`
- `spogo library albums remove <id|url...>`
- `spogo library artists list [--limit N] [--after <artist-id>]`
- `spogo library artists follow <id|url...>`
- `spogo library artists unfollow <id|url...>`
- `spogo library playlists list [--limit N]`

### playlists

- `spogo playlist create <name> [--public] [--collab]`
- `spogo playlist add <playlist> <track...>`
- `spogo playlist remove <playlist> <track...>`
- `spogo playlist tracks <playlist> [--limit N]`

### devices

- `spogo device list`
- `spogo device set <name|id>`

## Output contract

- stdout: primary results; human or machine modes.
- stderr: warnings/errors/logs.
- `--plain`: stable, line-oriented, tab-separated fields.
- `--json`: stable, documented keys per command.

## Engines

- `connect`: internal connect-state endpoints for playback; GraphQL for search/info.
- `web`: Web API endpoints; playback auto-fallback to connect when rate limited.

## Exit codes

- `0` success
- `1` generic failure
- `2` invalid usage/validation
- `3` auth/cookies missing or invalid
- `4` network/timeouts

## Config / env

- Env prefix: `SPOGO_`
- Precedence: flags > env > config
- Secrets: never via flags; use browser cookies only.
- Overrides:
  - `SPOGO_TOTP_SECRET_URL` (http(s) or `file://...`)
  - `SPOGO_CONNECT_VERSION` (connect playback client version)

## Examples

- `spogo auth import --browser chrome`
- `spogo search track "weezer" --limit 5 --plain`
- `spogo play spotify:track:7hQJA50XrCWABAu5v6QZ4i`
- `spogo device list --json`
- `spogo playlist create "Road Trip" --public`
