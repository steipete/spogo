---
title: Output
description: "spogo's output contract — human, plain, and JSON modes; stdout vs stderr; color and verbosity controls."
---

# Output

spogo follows a strict separation: **stdout** carries data, **stderr** carries logs and errors. Pipes always work — `spogo X | tool Y` never gets contaminated with progress bars or color codes when the destination isn't a TTY.

## Three output modes

### Human (default)

Coloured, formatted, friendly. Tables, headings, dimmed metadata. What you want when you're at a terminal.

```bash
spogo status
```

Color is automatic when stdout is a TTY. It is disabled when:

- `--no-color` is passed.
- `NO_COLOR` is set (any value).
- `TERM=dumb`.
- stdout is not a TTY (piped, redirected).

### `--plain`

Line-oriented, tab-separated, **stable**. Designed for `awk`, `cut`, `xargs`, and shell pipelines:

```bash
spogo search track "weezer" --limit 3 --plain
# spotify:track:7hQJA50XrCWABAu5v6QZ4i  Say It Ain't So     Weezer
# spotify:track:0sf12qNH5qcw8qpgymFOqD  Buddy Holly         Weezer
# spotify:track:4PTG3Z6ehGkBFwjybzWkR8  Undone — The Sweater Song   Weezer
```

Field order per command is documented in the [Spec](spec.md). Tabs are the only delimiter; values containing tabs are escaped.

### `--json`

Structured, **stable** keys. Use `jq` (or any JSON tool) downstream:

```bash
spogo status --json | jq -r '.item.name + " — " + (.item.artists|map(.name)|join(", "))'
```

JSON shapes match the [Spec](spec.md). Fields may be added; existing keys are not renamed or removed without a major version bump.

## Verbosity

| Flag | Effect |
| --- | --- |
| (default) | Normal: prints results to stdout, errors to stderr. |
| `-q`, `--quiet` | Suppress non-essential stderr output. |
| `-v`, `--verbose` | Extra context on stderr (engine choices, fallbacks, timings). |
| `-d`, `--debug` | Everything `-v` plus HTTP request/response details. |

Debug mode is the right escalation when something is misbehaving — see [Troubleshooting](troubleshooting.md).

## Stdout vs stderr

- **stdout**: command results only.
- **stderr**: warnings, errors, debug logs, prompts.

This is invariant — every spogo command in every mode follows it. That means `2>/dev/null` mutes diagnostic noise without losing data, and `>file.json` always captures clean output.

```bash
spogo library tracks list --json --limit 100 > tracks.json 2>/dev/null
```

## No prompts in pipelines

When stdin is not a TTY, spogo never prompts — commands that would normally ask for input return an error instead. Force prompts off explicitly with `--no-input`.

```bash
spogo auth paste --no-input < cookies.txt
```

## Color in CI

CI logs usually want color stripped. spogo respects `NO_COLOR` and detects non-TTY stdout, so the default behavior is correct in GitHub Actions, GitLab CI, etc. — no flag needed.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | Generic failure |
| `2` | Invalid usage / validation error |
| `3` | Auth / cookies missing or invalid |
| `4` | Network / timeout |

Use these in scripts:

```bash
if ! spogo auth status >/dev/null 2>&1; then
  case $? in
    3) echo "Need to re-import cookies" >&2 ;;
    *) echo "Auth check failed" >&2 ;;
  esac
  exit 1
fi
```

## Examples

```bash
# Pipe-friendly: just the URI
spogo search track "weezer" --limit 1 --plain | awk '{print $1}'

# Capture full JSON, pluck a field
spogo status --json | jq -r '.item.id'

# Mute stderr noise but keep stdout
spogo library tracks list --json 2>/dev/null > tracks.json

# Force color even in a pipe (rare; for fancy renderers)
spogo --no-color=false status | bat -l ansi
```
