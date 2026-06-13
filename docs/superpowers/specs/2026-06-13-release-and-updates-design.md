# nvim-quest — Release Engineering & Auto-Update Design

**Date:** 2026-06-13
**Status:** Approved for implementation

Ports three proven patterns from
[StrangeNoob/speed-test-cli](https://github.com/StrangeNoob/speed-test-cli) into
nvim-quest, adapted for a TUI game.

## Goals

1. **Version command** — `nvim-quest version` and `--version`, with version/commit/date
   injected at build time.
2. **CI/CD** — GitHub Actions: CI on push/PR, GoReleaser release on `v*` tags.
3. **Auto-update** — `nvim-quest update` subcommand plus a throttled launch-time notice
   on the title screen.

## 1. Build info & version command

- `main.go` declares `version = "dev"`, `commit = "none"`, `date = "unknown"` and passes
  them to `cmd.Execute(version, commit, date)`.
- `cmd` stores the build info; `rootCmd.Version = version` enables `--version`.
- A `nvim-quest version` subcommand prints: `nvim-quest <version> (commit <commit>, built <date>)`.
- GoReleaser injects the real values via `-ldflags -X main.version=… -X main.commit=… -X main.date=…`.

## 2. CI/CD

- `.github/workflows/ci.yml` — on push/PR to `main`: `go vet ./...`, `go build ./...`,
  `go test ./...`, `go test -race ./...`. (No `-short`; nvim-quest has no network tests.)
- `.github/workflows/release.yml` — on tag `v*`: GoReleaser (`goreleaser-action@v6`,
  `contents: write`).
- `.goreleaser.yaml` (v2): `project_name: nvim-quest`; build `main: .`, `binary: nvim-quest`,
  `CGO_ENABLED=0`, `-trimpath`, ldflags inject version/commit/date; goos linux/darwin/windows,
  goarch amd64/arm64, ignore windows/arm64; archives tar.gz (zip on windows) bundling
  `README.md` + `LICENSE`; `checksums.txt`; changelog from GitHub excluding docs/test/ci/chore;
  `prerelease: auto`.

## 3. Auto-update

New package `internal/update` (adapted from the reference):

- **`version.go`** — `Newer(current, latest) bool` and `upToDate(current, latest) bool`
  via `golang.org/x/mod/semver`. `dev`/unparseable current is "not up to date" so an
  explicit update still upgrades it; `Newer` returns false for `dev`.
- **`cache.go`** — `Cache{LastCheck, LatestVersion}`, `DefaultCachePath()` =
  `~/.nvim-quest/update-check.json` (same dir as the save file), `Load`/`Save` (best-effort),
  `Due(c, now, interval)`, `CheckInterval = 24h`.
- **`decide.go`** — `ShouldCheck(noFlag bool, env, version string) bool`: false when the
  flag is set, the env var `NVIM_QUEST_NO_UPDATE_CHECK` is set, or version is `dev`.
- **`remote.go`** — `Latest(ctx) (string, error)` and `Apply(ctx, current) (string, error)`
  via `creativeprojects/go-selfupdate`, slug `StrangeNoob/nvim-quest`, validated against
  `checksums.txt`. `Apply` replaces the running binary atomically.

**`nvim-quest update` subcommand** (`cmd/update.go` + a shared `runUpdate`): always checks
GitHub and self-updates, printing the outcome (up-to-date / upgraded / error with guidance)
to stderr.

**Launch-time notice (TUI):**
- `cmd/root.go` gains a persistent `--no-update-check` flag.
- The no-arg run computes whether to check (`update.ShouldCheck`) and passes the current
  version + a "check enabled" flag into `ui.New`.
- `ui.Model.Init()` returns a `tea.Cmd` that, when checking is enabled, reads the cache;
  if a real query is due it calls `update.Latest` (10s timeout) and saves the cache,
  otherwise uses the cached value. It returns an `updateCheckedMsg{latest string}`.
- `Update` stores `latest` in the model; the **title screen** renders
  `✨ vX available — run nvim-quest update` when `update.Newer(current, latest)`.
- Network/cache work runs in the `tea.Cmd` goroutine — the title screen shows instantly and
  the notice appears a moment later (or never, when throttled/up-to-date/offline).

## Testing

- `internal/update`: `version_test` (Newer/upToDate cases), `cache_test` (Due, round-trip,
  missing/corrupt file), `decide_test` (ShouldCheck matrix). A network `remote` test is
  build-tagged/skipped by default (like the reference).
- `cmd`: `version` output contains the injected values; `--no-update-check` is a valid flag.
- `ui`: the title view shows the notice when the model has a newer `latest`, and does not
  when `latest` is empty/older; `Init` returns no-op when checking is disabled (dev).

## Dependencies

- `github.com/creativeprojects/go-selfupdate`
- `golang.org/x/mod`

## Out of scope

Homebrew tap / install script, signed releases, and in-game (TUI) self-update without
re-launch — the `update` subcommand is the supported upgrade path.
