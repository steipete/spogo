---
title: Auth
description: "How spogo authenticates with Spotify using your browser cookies — import, paste, status, troubleshooting."
---

# Auth

spogo does not use the Spotify Developer API. It reads the cookies your browser already has for `open.spotify.com` and uses them to fetch a web access token. That means **no app registration, no client ID, no redirect URI** — just log in to Spotify in your browser, then import.

The cookie machinery comes from [steipete/sweetcookie](https://github.com/steipete/sweetcookie).

## What spogo needs

The minimum cookies for authentication:

- `sp_dc` — required. Long-lived web session cookie.
- `sp_key` — optional, helps with rotation.
- `sp_t` — recommended for `connect` engine playback control.

These cookies live in your browser's cookie store and rotate on their own; spogo refreshes its cached access token using them as needed.

## Importing from a browser

```bash
spogo auth import --browser chrome
```

Defaults: Chrome + `Default` profile + `spotify.com` domain. Cookies are stored under your config directory keyed by profile.

### Pick a different browser

```bash
spogo auth import --browser brave
spogo auth import --browser edge
spogo auth import --browser firefox
spogo auth import --browser safari
```

### Pick a non-default profile

Chrome / Brave / Edge keep profiles in directories like `Default`, `Profile 1`, `Profile 2`. Pass the directory name:

```bash
spogo auth import --browser chrome --browser-profile "Profile 1"
```

### Use a specific cookie store file

If you have an exported cookie jar already:

```bash
spogo auth import --cookie-path /path/to/cookies.sqlite
```

### Limit the cookie scope

```bash
spogo auth import --domain spotify.com
```

When the browser-store read returns nothing, spogo now surfaces the underlying warning (locked keychain, missing profile, decryption failure) instead of just printing `no cookies found`.

## Manual paste (WSL fallback)

If WSL cookie decryption is broken, or you need to copy cookies from a Chromium DevTools session, paste the values straight in:

1. In Chrome, open DevTools → Application → Cookies → `https://open.spotify.com`.
2. Copy the values for `sp_dc` (required), `sp_key` (optional), `sp_t` (recommended).
3. Run:

```bash
spogo auth paste
```

spogo prompts for each cookie. To skip the prompts (CI, scripts):

```bash
printf '%s\n%s\n' "sp_dc=..." "sp_t=..." | spogo auth paste --no-input
```

Other paste flags:

- `--cookie-path <file>` — write the resulting cookie jar to a custom path.
- `--domain <suffix>` — override the cookie domain (default `spotify.com`).
- `--path <path>` — override the cookie path (default `/`).

## Status & clearing

```bash
spogo auth status        # which profile, when imported, what cookies exist
spogo auth clear         # delete the stored cookies for the current profile
```

`auth status` does not call Spotify; it only inspects the local store. To verify cookies actually work, run any read command:

```bash
spogo status
spogo search track "test" --limit 1
```

A `401`/`403` from those means the cookies are stale — re-import.

## Where cookies are stored

- macOS: `~/Library/Application Support/spogo/<profile>/cookies.json`
- Linux: `~/.config/spogo/<profile>/cookies.json`
- Windows: `%APPDATA%\spogo\<profile>\cookies.json`

`<profile>` defaults to `default` — override with `--profile <name>` or `SPOGO_PROFILE`.

## Multiple accounts

Use profiles to keep multiple Spotify logins side by side:

```bash
spogo --profile work auth import --browser chrome --browser-profile "Profile 1"
spogo --profile personal auth import --browser chrome --browser-profile "Default"

spogo --profile work status
spogo --profile personal play spotify:track:...
```

Set the default for a shell with `export SPOGO_PROFILE=work`.

## Troubleshooting

- **"no cookies found"** — pass `--browser-profile`, double-check you're logged in to `open.spotify.com` in that browser, and check the warning spogo prints (it now surfaces the real reason).
- **Locked keychain (macOS Chrome)** — unlock the login keychain, then re-run `auth import`.
- **WSL Chrome** — cookie decryption is unreliable; use [paste](#manual-paste-wsl-fallback).
- **Auth works locally but not in CI** — copy the cookie jar file (the path printed by `auth status`) into the CI runner before running spogo.

See [Troubleshooting](troubleshooting.md) for more.
