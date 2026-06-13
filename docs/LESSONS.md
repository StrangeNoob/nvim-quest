# Lessons: schema, authoring, and roadmap

Lessons are plain JSON in [`assets/lessons/`](../assets/lessons/), embedded into the
binary at build time. Adding a lesson needs **no code changes** — drop in a file and the
content-integrity test picks it up.

## File naming

```
act<ACT>-<ORDER>-<slug>.json      e.g. act2-06-the-deletion-pits.json
```

`ACT` is 1–3 and `ORDER` is the global lesson order (lessons are sorted by `(act, order)`).
Keep `order` unique and increasing across the whole game.

## Schema

```jsonc
{
  "id": "act2-06-the-deletion-pits",   // unique lesson id
  "act": 2,                            // 1..3
  "order": 6,                          // global sort order
  "title": "The Deletion Pits",
  "story": "Flavor text shown under the lesson title.",
  "challenges": [
    {
      "id": "a2l6c1",                  // unique across ALL challenges & boss steps
      "intro": "What the player must do, in-world.",
      "buffer": ["the line(s)", "of starting text"],
      "cursor": [0, 4],               // [row, col]; col may equal len(line) for append
      "goal": { "type": "bufferEquals", "lines": ["the goblin lurks"] },
      "par": 2,                        // keystrokes for 3 stars (must be >= 1)
      "xp": 60,                        // base XP awarded on first clear (must be > 0)
      "hint": "dw deletes a word.",   // REQUIRED and non-empty — shown by the [?] key
      "newKeys": ["dw"],              // optional: keys introduced here (display only)
      "allowedKeys": ["h","j","k","l","w","b","x","d"]  // optional whitelist; omit = allow all
    }
  ],
  "boss": {                            // optional; only on each act's FINAL lesson
    "name": "The Gravewright",
    "taunt": "Shown under the boss name.",
    "timeLimitSec": 90,               // >= 30
    "xp": 250,                         // > 0
    "steps": [                         // ordered; each is a Challenge (no par/xp needed,
      {                                // but hint IS required — see the integrity test)
        "id": "a2boss1",
        "intro": "Cut the doubled curse letter.",
        "buffer": ["the ccurse holds"], "cursor": [0, 4],
        "goal": { "type": "bufferEquals", "lines": ["the curse holds"] },
        "hint": "x deletes the character under the cursor."
      }
    ]
  }
}
```

### Goal types

| Type | Fields | Passes when |
| --- | --- | --- |
| `cursorOnWord` | `word` | the cursor rests on `word` |
| `bufferEquals` | `lines` | the buffer matches `lines` exactly |
| `lineDeleted` | `line` | `line` no longer exists in the buffer |
| `wordDeleted` | `word` | `word` appears nowhere in the buffer |
| `containsText` | `text` | some line contains `text` |
| `searchMatchActive` | `term` | a search ran for `term` and the cursor sits on a match |

The payload field must be non-empty (an empty `word`/`term`/etc. never matches — guarded
by both the engine and the integrity test).

### `allowedKeys`

Lists the keys the player may use in Normal mode for that challenge. Anything else costs a
heart. This is how a lesson keeps the player focused on what it's teaching — e.g. the
first motion lesson allows only `h j k l`. Build the set as *"everything taught so far"*
plus the new keys. `Esc`, `?`, and quit are always allowed; insert/search typing is never
gated. Omit `allowedKeys` (boss steps do) to allow everything.

## Authoring rules (enforced by tests)

`internal/content/loader_test.go` and `solvable_test.go` will fail the build unless:

1. Every lesson has a unique `id`, an act in 1–3, and `order` increasing across the game.
2. Every challenge & boss-step `id` is globally unique.
3. Every `cursor` is in bounds (`col` may equal the line length for append-at-end).
4. Every `goal.type` is valid and its payload field is non-empty.
5. **Every challenge and boss step has a non-empty `hint`** (the `[?]` key renders it; an
   empty hint shows the player nothing).
6. Regular challenges have `par >= 1` and `xp > 0`; bosses have `timeLimitSec >= 30`,
   `xp > 0`, and at least one step.
7. The act's final lesson (and only it) carries a `boss`.
8. **Every challenge is solvable at par.** `solvable_test.go` plays an authored optimal
   key sequence through the real engine and asserts the goal is met within par. When you
   add a challenge, add its solution there too.

Run `go test ./internal/content/` after editing content.

## Worked example

To add a lesson teaching `e` (jump to end of word) — once the engine supports `e`:

1. Create `assets/lessons/act1-03b-word-ends.json` (renumber later lessons' `order` if
   you want it to slot in mid-act, or append it).
2. Give it one or two `cursorOnWord` challenges with `allowedKeys` including `e`.
3. Add each challenge's optimal solution to `solvable_test.go`.
4. `go test ./internal/content/` — green means it's wired in and winnable.

---

# Roadmap: future lessons

The current game (v2) covers the daily-driver core. The roadmap below is grouped by the
**engine work** each batch needs, because content can only teach what the simulator can
do. Phases mirror the original design spec
([`docs/superpowers/specs/2026-06-13-nvim-quest-v2-design.md`](superpowers/specs/2026-06-13-nvim-quest-v2-design.md)).
Curriculum source: [Learn-Vim-and-NeoVim](https://github.com/rcallaby/Learn-Vim-and-NeoVim).

### Phase "Solid" — extends the existing three acts

These slot into the current acts and need only small, localized engine additions.

**Shipped (batch 1 — the Editing pack, in Act II):** `o`/`O` (Opening Lines), `P` (Echo
Chamber), `ciw`/`diw` (The Inner Cut), `u`/`Ctrl-r` (Rewind).

Still to do:

| Lesson idea | Teaches | Suggested act | Engine work |
| --- | --- | --- | --- |
| **The Word's Edge** | `e` (end of word) | I | small: an `e` motion (mirror of `w`) |
| **The Marksman** | `f` / `t` / `;` / `,` (find/till char on a line) | III | small: intra-line char search |
| **The Great Substitution** | `:%s/old/new/g` (and ranges/flags) | III | medium: command-line mode + substitute parser |

### Phase "Everything" — likely a new Act IV (and beyond)

Bigger features that probably deserve their own world (e.g. **Act IV · The Archives**, a
library/clockwork theme for power-user automation).

| Lesson idea | Teaches | Engine work |
| --- | --- | --- |
| **The Selection** | Visual mode `v` + `d`/`y`/`c` on a selection | medium: a visual mode with an anchor + selection range |
| **Block Party** | Visual block `Ctrl-v`, column insert `I`/`A` | larger: rectangular selection & multi-line column edits |
| **Waypoints** | Marks `m{a-z}` and jumps `` `{a-z} ``/`'{a-z}` | medium: a per-buffer mark table |
| **The Registers** | Named registers `"a`–`"z` for yank/paste | medium: a register map; thread it through yank/delete/paste |
| **The Macro Forge** | Record `q{reg}` … `q`, replay `@{reg}`, `@@` | larger: record a keystroke stream into a register and replay it |

### How a new act fits

The world map, unlock logic, palettes, and boss flow are all data-driven by `act`/`order`
and the per-act palette map in `internal/ui/styles.go`. To add **Act IV** you would:

1. Add an entry to `palettes` and a name to `actName()` in `internal/ui/styles.go`.
2. Add `act4-*.json` lesson files (final one carrying a boss) — and their solutions in
   `solvable_test.go`.
3. Relax the act-range checks (currently `1..3`) in `internal/content/loader_test.go`.

Everything else (map rendering, unlocks, results, badges) already generalizes over the
number of acts and lessons.

### Non-lesson ideas (from the spec's out-of-scope list)

Daily challenge, leaderboard, a user config file, custom/community lessons, and real
Neovim integration — all deliberately out of scope for now, listed here so the intent
isn't lost.
