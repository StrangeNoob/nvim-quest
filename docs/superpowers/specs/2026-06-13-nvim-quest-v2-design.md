# nvim-quest v2 — Design Spec

**Date:** 2026-06-13
**Status:** Approved for implementation planning
**Decision:** Full rewrite. The game is the first-class design; good ideas from v1 (Vim simulator, goal validators, JSON lessons) are re-implemented cleanly.

## Vision

Transform nvim-quest from a quiz-shaped tutorial into a real game: a three-act journey
through themed worlds, played with **real-time Vim keystrokes** (no Enter-to-submit),
with golf-style par scoring, combo streaks, hearts, and timed boss fights.

Content is sourced from the Learn-Vim-and-NeoVim curriculum
(https://github.com/rcallaby/Learn-Vim-and-NeoVim).

## Scope

**MVP (this spec):** "Light touch" content — the v1 command set, re-themed into three
acts, plus change operations (`cw`, `cc`) and count prefixes (`3w`, `2dd`).
10 lessons + 3 bosses.

**Later phases (out of scope, design accommodates them):**
- **Solid:** `e`, `o`/`O`, `P`, `ciw`, `f`/`t`, `:%s/old/new/g`, dedicated undo/redo lesson
- **Everything:** visual mode, visual block, marks, registers, macros

## Game flow & screens

One Bubble Tea program, one root model that switches between five screens:

1. **Title** — ASCII logo, menu: Continue / World Map / Stats / Quit.
2. **World Map** — the three-act journey tree. Lesson states: ✓ complete (with best
   stars), ▶ current, 🔒 locked. Lessons unlock sequentially; an act's boss unlocks
   the next act.
3. **Room** (challenge) — story intro line, text buffer with block cursor, HUD:
   mode indicator (`-- NORMAL --` / `-- INSERT --`), hearts ♥♥♥, combo ⚡xN,
   keystroke count vs par (`4/5⛳`), pending operator/count display.
4. **Boss** — Room plus a draining countdown bar and a multi-step objective checklist.
5. **Results** — stars earned, XP gained (with combo multiplier breakdown),
   badge unlocks, continue → next room or back to map.

**CLI:** `nvim-quest` (no args) launches the game at the Title screen.
Subcommands kept: `stats` (print progress summary), `reset` (confirm + wipe progress).
v1's `lesson` and `practice` subcommands are dropped for MVP (the world map replaces
them); `practice` may return in a later phase.

## Architecture

```
main.go
cmd/              cobra: root (launch game), stats, reset
internal/
  engine/         real-time Vim emulator: buffer, cursor, modes,
                  keystroke state machine, undo, yank register, search
  game/           mechanics: stars/par, combo, hearts, XP/levels, badges
  content/        lesson + act data models, go:embed JSON loader
  progress/       versioned save file ~/.nvim-quest/progress.json
  ui/             screens, HUD components, per-act lipgloss palettes
assets/lessons/   JSON content, one file per lesson
```

Dependencies: bubbletea, lipgloss, cobra (same stack as v1; bubbles as needed).
Lessons are embedded via `go:embed`, fixing v1's "must run from project root" defect.

## Engine

### API

```go
type Event struct {
    Kind    EventKind // Moved, Inserted, Deleted, Yanked, Pasted, Undone,
                      // ModeChanged, OperatorPending, CountPending,
                      // SearchUpdated, SearchJumped, InvalidKey, NoOp
    // ... details as needed by the UI (e.g., deleted text, new mode)
}

func (s *Simulator) Press(key string) Event
```

Each keystroke immediately mutates simulator state and returns an event the UI
reacts to (flash on success, shake + heart loss on InvalidKey).

### State machine

Internal state: `mode` (Normal / Insert / Search), `pendingOp` (`d`/`c`/`y`),
`pendingCount`, `pendingG` (for `gg`), buffer, cursor, yank register
(line-wise vs character-wise), undo stack (snapshot before each mutation),
search query + last match.

- **Normal mode:** `h j k l w b 0 $ G x i a u p` single keys; `g` arms pendingG
  (`gg` jumps to top); `d`/`c`/`y` arm pendingOp, then a motion or doubled key
  completes it (`dw`, `dd`, `cw`, `cc`, `yy`); digits accumulate pendingCount and
  apply to the following motion/operation (`3w`, `2dd`, `4x`).
- **Insert mode:** printable characters insert at cursor; `Esc` returns to Normal
  (cursor moves back one column, like Vim). Entered via `i` (insert) or `a` (append).
  `cw`/`cc` delete then enter Insert.
- **Search mode:** `/` opens a live prompt; typed chars build the query; `Enter`
  jumps to the first match; `Esc` cancels; back in Normal mode `n` repeats forward.
- **Invalid key:** a key that is illegal in the current state, or not in the
  challenge's allowed-key set, emits `InvalidKey`. Keys that are always allowed and
  never penalized: `Esc`, `?` (hint toggle), `Ctrl+C` / quit chrome.

MVP command set: `h j k l w b 0 $ gg G i a x dw dd cw cc yy p u / n Esc`
plus count prefixes.

### Goal validators

Six goal types, checked after every event:

| Type | Passes when |
|---|---|
| `cursorOnWord` | cursor rests on the given word |
| `bufferEquals` | buffer matches expected lines exactly |
| `lineDeleted` | the given line no longer exists |
| `wordDeleted` | the given word is gone from the buffer |
| `containsText` | buffer contains the given text |
| `searchMatchActive` | search executed for term, cursor on a match |

Boss steps each carry one goal; steps complete in order.

## Content model

One JSON file per lesson in `assets/lessons/`, named `act<N>-<NN>-<slug>.json`:

```json
{
  "id": "act1-03-way-of-the-word",
  "act": 1,
  "order": 3,
  "title": "Way of the Word",
  "story": "Sensei: \"The wise cursor moves by words, not steps.\"",
  "challenges": [
    {
      "id": "a1l3c1",
      "intro": "Reach the temple in as few strides as you can.",
      "buffer": ["the path to the temple is long"],
      "cursor": [0, 0],
      "goal": { "type": "cursorOnWord", "word": "temple" },
      "par": 4,
      "hint": "w jumps forward one word at a time.",
      "newKeys": ["w", "b"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b"]
    }
  ],
  "boss": {
    "name": "Sensei's Trial",
    "timeLimitSec": 60,
    "taunt": "Show me everything you have learned, student.",
    "steps": [
      { "intro": "...", "buffer": ["..."], "cursor": [0, 0],
        "goal": { "type": "cursorOnWord", "word": "..." } }
    ]
  }
}
```

- `allowedKeys` defaults to "all keys taught so far in the journey" when omitted.
- `boss` appears only on each act's final lesson.
- A content-integrity test validates every JSON at build time.

### MVP lesson list

**Act I · The Cursor Dojo** (calm greens/cream — student learning the way)
1. The Two Stances — modes, `i`, `Esc`
2. First Steps — `h j k l`
3. Way of the Word — `w`, `b`
4. The Great Leaps — `0`, `$`, `gg`, `G` → **Boss: Sensei's Trial** (timed motion gauntlet)

**Act II · The Motion Crypts** (ember amber on dark — dungeon crawl)
5. Hall of Insertion — `i`, `a`
6. The Deletion Pits — `x`, `dw`, `dd`
7. Echo Chamber — `yy`, `p`
8. The Shapeshifter — `cw`, `cc` *(new vs v1)* → **Boss: The Gravewright** (timed editing)

**Act III · The Neon Grid** (magenta/cyan — cyberpunk infiltration)
9. Trace Evasion — `/`, `n`
10. Power Surge — counts: `3w`, `2dd`, `4x` *(new vs v1)* → **Final Boss: The Grid Core** (everything, tight timer)

## Mechanics

- **Stars (par golf):** solve in ≤ par keystrokes → ⭐⭐⭐; ≤ par+2 → ⭐⭐; any solve
  → ⭐. Best stars per challenge persist; replays can improve them. Keystrokes that
  count: every key fed to the engine except always-allowed chrome keys.
- **Hearts:** 3 per room. `InvalidKey` costs one heart. At zero hearts the room
  resets (buffer, keystroke count) — no XP penalty, but the combo breaks.
- **Combo:** consecutive clean room clears (no heart lost) raise the multiplier
  x1 → x5 (cap). XP awarded = base XP × multiplier. Combo persists across rooms and
  lessons within a session; any heart loss resets it to x1, and it does not survive
  app restart.
- **XP & levels:** each challenge has base XP; level thresholds follow a simple
  quadratic-ish curve (e.g., level N needs 100·N XP beyond the previous). Level shown
  on map and HUD.
- **Badges:** act completion (3), plus: *First Steps* (first challenge), *Birdie*
  (beat par 10 times), *Untouchable* (clear an act losing no hearts),
  *Grid Breaker* (finish the final boss).
- **Boss fights:** countdown driven by `tea.Tick` (1s granularity, smooth bar).
  Steps complete in order; timer expiry → dramatic fail screen → retry from step 1.
  Success → act-complete celebration + badge.

## Progress persistence

`~/.nvim-quest/progress.json`, version 2:

```json
{
  "version": 2,
  "xp": 340,
  "level": 2,
  "stars": { "a1l3c1": 3 },
  "completed": ["a1l3c1"],
  "badges": ["First Steps"],
  "lastLesson": "act1-03-way-of-the-word"
}
```

On load: missing file → fresh journey. `version != 2` (or unparsable) → back up the
old file to `progress.json.v1.bak` and start fresh. Save after every challenge
completion and boss result; save errors surface as a non-fatal warning in the UI.

## Visual identity

- Per-act lipgloss palettes (Dojo green/cream, Crypts amber/dark, Grid magenta/cyan);
  the title screen and map tint by current act.
- Block cursor rendered by inverting the cell; pending operator shown beside the mode
  indicator like real Vim (`d`, `3`).
- Success: buffer flashes green for ~300ms (tick-based) before Results.
- Heart loss: HUD shake/flash red one frame.
- Boss timer bar pulses faster as it drains.
- Graceful degradation: no animation is load-bearing; everything readable in a plain
  80×24 terminal.

## Testing

- **Engine:** table-driven keystroke-sequence tests — `[]keys → expected buffer,
  cursor, mode` — covering every MVP command, operator combos, counts, undo, and
  invalid-key events.
- **Validators:** one test per goal type.
- **Content integrity:** all embedded lesson JSONs parse; references are sane
  (acts 1–3, unique IDs, par ≥ 1, goals well-formed); a solver-style check that each
  challenge's `allowedKeys` can reach its goal is a stretch goal.
- **Mechanics:** star thresholds, combo multiplier math, heart/reset flow, XP/level
  curve, badge triggers.
- **Progress:** round-trip save/load, v1-file backup behavior.
- **UI:** model-level Update tests for screen transitions (map → room → results);
  no terminal-emulator tests in MVP.

Single command: `go test ./...`.

## Error handling

- Save failures: non-fatal UI warning, game continues.
- Corrupt progress: back up and start fresh (never crash on bad JSON).
- Terminal too small (< 80×24): show a friendly "resize me" screen instead of
  broken layout.
- Embedded content failing to parse is a programmer error caught by tests; loader
  panics with a clear message rather than rendering a broken game.

## Out of scope (MVP)

- Visual mode, registers, macros, marks, `:%s`, `f`/`t` (later phases)
- Practice/daily-challenge modes, leaderboards, config files, custom lessons
- Real Neovim integration
