# Architecture

nvim-quest is a single Go binary built on [Bubble Tea](https://github.com/charmbracelet/bubbletea)
(TUI), [Lipgloss](https://github.com/charmbracelet/lipgloss) (styling), and
[Cobra](https://github.com/spf13/cobra) (CLI). The module is `nvim-quest`, Go 1.24.

## Package layout

```
main.go                  entry point → cmd.Execute()
cmd/                     Cobra commands
  root.go                no-arg command: loads content + progress, runs the TUI
  stats.go               prints a progress summary
  reset.go               confirm + delete the save file
assets/
  assets.go              //go:embed of lessons/*.json
  lessons/*.json         all game content (10 lessons, 3 bosses)
internal/
  engine/                the Vim emulator (no UI, no content deps beyond Goal)
    engine.go            Simulator, modes, Press dispatch, normal-mode motions, counts
    edit.go              insert mode, x, operators (dw/dd/cw/cc), yank/paste, undo
    search.go            search mode (/) and n, with wraparound
    goal.go              Goal type + Met() validators, WordAtCursor
  content/               lesson data + loader
    model.go             Lesson / Challenge / Boss structs
    loader.go            All() — parse the embedded JSON, sort by (act, order)
  game/                  pure scoring math (stars, combo, XP/levels, badges)
  progress/              versioned save file (~/.nvim-quest/progress.json) + v1 backup
  ui/                    Bubble Tea front end (the only package that imports all others)
    app.go               root Model, screen enum, Update/View dispatch
    keys.go              normalizeKey (Bubble Tea key names → engine vocabulary)
    styles.go            per-act color palettes
    welcome.go           first-launch welcome screen
    title.go             title menu + stats screen
    worldmap.go          world map + unlock logic
    room.go              gameplay: challenges & bosses, HUD, tick timer, scoring
    results.go           results / act-complete / failure screen
```

### Dependency direction

```
cmd ──▶ ui ──▶ engine
        │ │     ▲
        │ ├────▶ content ──▶ engine (only for the Goal type)
        │ ├────▶ game
        │ └────▶ progress
        └──────▶ content, progress
```

`engine` is the leaf — it has no dependency on UI, content, scoring, or persistence.
`content` depends on `engine` only to embed `engine.Goal` in its JSON model. `ui` is the
composition root that wires everything together. This keeps the simulator and the scoring
math independently testable.

## The engine

The engine is a small, deterministic Vim emulator. The whole contract is one method:

```go
func (s *Simulator) Press(key string) Event
```

Each keystroke mutates the simulator's state (buffer, cursor, mode, registers, undo
stack) and returns an `Event` describing what happened (`EvMoved`, `EvEdited`,
`EvModeChanged`, `EvPending`, `EvSearchJumped`, `EvInvalid`, `EvNone`). The UI reacts to
the event (e.g. `EvInvalid` costs a heart) and, after every keystroke, asks the active
`Goal` whether the challenge is solved.

**Modes:** Normal, Insert, Search. `Press` dispatches by mode.

**Pending state** holds an in-progress command: a count prefix (`3`), an operator
(`d`/`c`/`y`), or the first `g` of `gg`. `Pending()` renders it for the HUD (e.g. `3d`).

**AllowedKeys** is an optional per-challenge whitelist. In Normal mode, a key outside the
set returns `EvInvalid` (and the UI docks a heart). This is how a lesson restricts the
player to the keys it's teaching. `Esc`, `?`, and quit chrome are never penalized; insert-
and search-mode typing is never gated.

**Goals** (`engine.Goal`) are the success conditions, decoded straight from lesson JSON.
`Goal.Met(sim)` is checked after every keystroke. The six validators:

| Type | Passes when |
| --- | --- |
| `cursorOnWord` | the cursor rests on the given (non-empty) word |
| `bufferEquals` | the buffer matches the given lines exactly |
| `lineDeleted` | the given line no longer exists |
| `wordDeleted` | the given word appears nowhere in the buffer |
| `containsText` | some line contains the given text |
| `searchMatchActive` | a search ran for the (non-empty) term and the cursor sits on a match |

> The engine indexes bytes, not runes — lesson content is ASCII. Word motions treat words
> as space-delimited (a deliberate simplification of Vim's word classes for teaching).

## Game loop & screens

`ui.Model` is one Bubble Tea model with a `screen` enum. `Update` routes a message to the
active screen's handler; `View` renders it. A `< 80×24` terminal shows a resize prompt
instead of a broken layout.

```
welcome ─▶ title ─┬─▶ map ─▶ room ─▶ results ─▶ (next room | boss | map)
                  ├─▶ stats              ▲
                  └─▶ (quit)             └─ boss (timed) ─▶ results
```

- **Welcome** shows only on first launch (`isFreshPlayer()` — no rooms cleared yet).
- **Room** feeds keystrokes to a fresh `engine.Simulator` built from the challenge's
  buffer/cursor/allowed-keys. On a clear it calls into `game` for stars/XP/combo/badges,
  writes to `progress`, and triggers a brief success flash before the results screen.
- **Boss** is a room with a `tea.Tick`-driven countdown and ordered multi-step goals.
  A generation counter (`tickGen`) invalidates stale ticks when a boss restarts; expiry
  routes to a failure result that retries from step 0.

## Scoring (`internal/game`)

Pure functions, no I/O:

- `Stars(keys, par)` → 3 / 2 / 1
- `NextCombo(c)` → `c+1` capped at 5; the UI resets it to 1 on any heart loss
- `XP(base, combo)` → `base * combo`
- `LevelForXP(xp)` → level from a quadratic-ish threshold curve (0, 100, 300, 600, …)
- badge constants + `ActBadge(act)` and `BirdieEarned(stars)`

## Persistence (`internal/progress`)

`Progress` (version 2) holds XP, level, per-challenge best stars, completed IDs, badges,
and the last lesson. `Load` never fails: a missing file starts fresh; a corrupt or
old-version file is renamed to `<path>.v1.bak` and a fresh save is started. `Save` writes
pretty-printed JSON, creating parent directories as needed. Save failures surface as a
non-fatal warning in the room view — the game keeps running.

## Content (`internal/content` + `assets/`)

Lessons are JSON, embedded via `//go:embed lessons/*.json`, so the installed binary
carries its content and runs from any directory. `All()` parses every file and sorts by
`(act, order)`. The schema and authoring guide are in [LESSONS.md](LESSONS.md).
