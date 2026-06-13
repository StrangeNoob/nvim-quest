# nvim-quest

**Learn Neovim through an epic three-act terminal quest — no Neovim required.**

nvim-quest teaches real Vim motions and edits by making you *use* them. You press
the actual keys (`h j k l`, `d w`, `c w`, `/` …) and a live buffer reacts instantly,
exactly like the real editor. Solve rooms in as few keystrokes as you can for stars,
build combo streaks, guard your hearts, and beat the clock in boss fights.

```
✨ YOUR JOURNEY ✨

ACT 1 · THE CURSOR DOJO
> ▶ The Two Stances
  🔒 First Steps
  🔒 Way of the Word
  🔒 The Great Leaps · boss: Sensei's Trial

ACT 2 · THE MOTION CRYPTS         ACT 3 · THE NEON GRID
  🔒 Hall of Insertion              🔒 Trace Evasion
  🔒 The Deletion Pits              🔒 Power Surge · boss: The Grid Core
  🔒 Echo Chamber
  🔒 The Shapeshifter · boss: The Gravewright
```

---

## Installation

**With Go (recommended)** — requires **Go 1.24+**. Lessons are embedded in the binary, so
it runs from anywhere:

```sh
go install github.com/StrangeNoob/nvim-quest@latest
nvim-quest
```

This installs to `$(go env GOPATH)/bin` (add it to your `PATH` if it isn't already).

**Prebuilt binary** — download the archive for your OS/arch from the
[Releases](https://github.com/StrangeNoob/nvim-quest/releases) page, extract it, and put
the `nvim-quest` binary on your `PATH`:

```sh
# example: macOS (Apple Silicon)
tar -xzf nvim-quest_*_darwin_arm64.tar.gz
sudo mv nvim-quest /usr/local/bin/
```

> Keep the binary named `nvim-quest` so `nvim-quest update` can self-replace it.

**From source** — clone the repo and:

```sh
go run .            # start the game
go run . stats      # print a progress summary
go run . reset      # erase all progress (asks to confirm)
```

## Play

```sh
nvim-quest           # start the game
nvim-quest version   # version, commit, build date
nvim-quest update    # self-update to the latest release
```

When a newer release exists, the title screen shows a `✨ vX available` notice (a
throttled, background check; disable it with `--no-update-check` or the
`NVIM_QUEST_NO_UPDATE_CHECK` env var). To upgrade, run `nvim-quest update` (or
`go install github.com/StrangeNoob/nvim-quest@latest`).

Your progress is saved to `~/.nvim-quest/progress.json`.

> The game needs an interactive terminal at least **80×24**. It runs on the
> [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

---

## How to play

You move through five screens: a **welcome** intro (first launch only), the **world
map**, a **room** (a single challenge), a **boss** fight, and a **results** screen.

**Navigation (menus & map):**

| Key | Action |
| --- | --- |
| `j` / `k` | move selection down / up (arrow keys are intentionally disabled — it's a Vim trainer) |
| `enter` | play the selected lesson |
| `esc` | back to the title |

**Inside a room you type real Vim keys.** The HUD shows everything you need:

```
-- NORMAL --   ♥ ♥ ♥    ⚡x2   keys 4 · par 5
[?] hint · [esc] back to map
```

- `-- NORMAL --` / `-- INSERT --` — the current mode (press `i` to type, `Esc` to leave)
- `♥ ♥ ♥` — **hearts**: a wrong/disallowed key costs one; lose all three and the room resets (no XP penalty)
- `⚡x2` — **combo**: clearing rooms without losing a heart multiplies your XP (up to ×5)
- `keys N · par M` — **par golf**: solve in `≤ par` keystrokes for ⭐⭐⭐, `≤ par+2` for ⭐⭐, otherwise ⭐
- `[?]` — toggle a hint at any time (free; never counts against par)

**Boss fights** end each act: a multi-step challenge against a draining timer. Run out
of time and you retry from the first step.

---

## The journey & what it teaches

| Act | World | Lessons | Vim covered |
| --- | --- | --- | --- |
| **I** | The Cursor Dojo | The Two Stances · First Steps · Way of the Word · The Great Leaps | modes & `i`/`Esc`, `h j k l`, `w`/`b`, `0` `$` `gg` `G` |
| **II** | The Motion Crypts | Hall of Insertion · The Deletion Pits · Echo Chamber · Opening Lines · The Inner Cut · Rewind · The Shapeshifter | `i`/`a`, `x` `dw` `dd`, `yy`/`p`/`P`, `o`/`O`, `ciw`/`diw`, `u`/`Ctrl-r`, `cw` `cc` |
| **III** | The Neon Grid | Trace Evasion · Power Surge | `/` search & `n`, count prefixes (`4w`, `3dd`, `4x`) |

Each act has a distinct color palette (dojo greens, crypt embers, neon magenta/cyan)
and ends with a timed boss. Lessons unlock sequentially; clearing an act's boss unlocks
the next act.

The curriculum is inspired by
[**Learn-Vim-and-NeoVim**](https://github.com/rcallaby/Learn-Vim-and-NeoVim) by Richard
Callaby — see [docs/LESSONS.md](docs/LESSONS.md) for the **roadmap of future lessons**.

---

## Development

```sh
go test ./...     # full suite (engine, content, game, progress, ui)
go vet ./...
gofmt -l .        # should print nothing
go build ./...
```

Highlights of the test suite:

- **`internal/engine`** — table-driven keystroke tests for every motion, operator, and edit
- **`internal/content`** — content-integrity test (every lesson parses, cursors are in
  bounds, goals/hints are populated) **plus a solvability test that plays an optimal
  solution through the real engine for every challenge** and asserts it's winnable at par
- **`internal/ui`** — model-level screen-transition tests (menus, rooms, bosses, results)

### Documentation

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — package layout, the engine, and data flow
- [docs/LESSONS.md](docs/LESSONS.md) — the lesson JSON schema, how to add a lesson, and the future-lessons roadmap
- `docs/superpowers/specs/` and `docs/superpowers/plans/` — the original design spec and implementation plan

---

## Adding a lesson

Lessons are plain JSON files in [`assets/lessons/`](assets/lessons/) — no code changes
needed. Drop in a file and the content-integrity test validates it automatically. The
schema and a worked example live in [docs/LESSONS.md](docs/LESSONS.md).

---

## License

Released under the [MIT License](LICENSE). The curriculum it's based on,
[Learn-Vim-and-NeoVim](https://github.com/rcallaby/Learn-Vim-and-NeoVim), is also
MIT-licensed.
