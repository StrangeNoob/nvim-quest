# nvim-quest — Editing Pack (Solid phase, batch 1)

**Date:** 2026-06-13
**Status:** Approved for implementation

Extends Act II with the everyday edits the MVP lacked. First batch of the "Solid"
roadmap phase.

## Engine additions (`internal/engine`)

1. **`o` / `O`** — open a new empty line below / above the cursor's line and enter
   Insert mode. New `openLine(below bool)` helper builds a fresh buffer slice
   (snapshot first); cursor goes to the new line, col 0.

2. **`P`** — paste the yank register *before* the current line (mirror of `p`, which
   pastes after). New `pasteBefore()`; cursor to `{row, 0}`. No-op (`EvNone`) on an
   empty register.

3. **`ciw` / `diw`** — change / delete *inner word* (text object). Adds an `i`
   text-object prefix to the operator state machine:
   - In `applyOperator`, when `pendingOp` is `d`/`c` and the key is `i`, set
     `pendingInner = true` and return `EvPending` (keeping `pendingOp`).
   - The next key `w` completes it: `deleteInnerWord()` removes the whole word the
     cursor sits on (its bounds, no trailing space). For `c`, then enter Insert.
   - `pendingInner` is reset by `clearPending`. Counts are not combined with text
     objects in this batch.
   - `deleteInnerWord`: if the cursor is on a space or past line end, no-op; otherwise
     delete `line[start:end]` where start/end are the word bounds, cursor → start.

4. **`Ctrl-r`** — redo. Adds a `redo []snapshot` stack:
   - `snapshot()` clears `redo` (a new edit invalidates the redo history).
   - `applyUndo()` pushes the current state onto `redo` before restoring from `undo`.
   - `applyRedo()` (new, bound to key `ctrl+r`) pops `redo`, pushes current onto `undo`,
     restores. `EvNone` when `redo` is empty.

All new normal-mode keys (`o O P ctrl+r`) are added to `pressNormal`; `i`/`w` already
exist (the text-object path reuses them via the operator state).

## Content — Act II renumbering

New lessons must precede the act's boss (the integrity test requires the boss to be the
act's last lesson). So Act II is renumbered:

| Order | Lesson | IDs | Status |
| --- | --- | --- | --- |
| 5 | Hall of Insertion | a2l5* | unchanged |
| 6 | The Deletion Pits | a2l6* | unchanged |
| 7 | Echo Chamber | a2l7c1, c2, **c3 (P)** | + 1 challenge |
| 8 | **Opening Lines** (`o`/`O`) | a2l8c1, c2 | new |
| 9 | **The Inner Cut** (`ciw`/`diw`) | a2l9c1, c2 | new |
| 10 | **Rewind** (`u` + `Ctrl-r`) | a2l10c1, c2 | new |
| 11 | The Shapeshifter (boss) | a2l11c1, c2 + a2boss1-4 | **moved** from order 8 |

Act III is unchanged (orders 9, 10; act 3 sorts after act 2 regardless of order value).

**Progress note:** moving the Shapeshifter renumbers its lesson id
(`act2-08-the-shapeshifter` → `act2-11-the-shapeshifter`) and its two challenge ids
(`a2l8c*` → `a2l11c*`). A local player who already cleared it will see that one lesson
(and its boss flag) as incomplete; all other progress is preserved. Acceptable — the game
shipped the same day. Boss step ids (`a2boss1-4`) are unchanged.

### New lesson content (exact buffers/pars finalized against the engine in
`solvable_test.go`; the values below are the design intent)

- **Echo Chamber c3 (`P`):** `["top","bottom"]`, cursor `[1,0]`; goal
  `bufferEquals ["top","top","bottom"]`; solution `k yy j P`; par 5.
- **Opening Lines c1 (`o`):** `["light"]`; goal `["light","dark"]`; solution `o dark`; par 5.
- **Opening Lines c2 (`O`):** `["second"]`; goal `["first","second"]`; solution `O first`; par 6.
- **The Inner Cut c1 (`ciw`):** `["the wrong word"]`, cursor on `wrong`; goal
  `["the right word"]`; solution `c i w right`; par 8.
- **The Inner Cut c2 (`diw`):** `["delete the cursed rune"]`, cursor on `cursed`; goal
  `wordDeleted "cursed"`; solution `d i w`; par 3.
- **Rewind c1 (`u`):** `["keep me","BURN THIS LINE"]`, cursor `[1,0]`; goal
  `lineDeleted "BURN THIS LINE"`; solution `dd`; hint teaches that `u` undoes a wrong
  delete. par 2.
- **Rewind c2 (`Ctrl-r`):** `["echo"]`; goal `["echo","echo"]`; solution `yy p`; hint
  teaches `u` removes the paste and `Ctrl-r` brings it back. par 3.

`newKeys` and `allowedKeys` extend the running Act II set with the lesson's new keys
(Rewind adds `u`, `ctrl+r`; Opening Lines adds `o`, `O`; Echo Chamber adds `P`).

## Testing

- **Engine:** table tests for `o`/`O` (open + insert + cursor), `P` (before, empty-register
  no-op), `diw`/`ciw` (inner-word bounds, cursor, mode), and undo→redo round-trips
  (including that a new edit clears redo).
- **Content integrity (`loader_test.go`):** assert 13 lessons and 3 bosses; existing
  cursor/goal/hint/par checks cover the new lessons.
- **Solvability (`solvable_test.go`):** add the 7 new challenge solutions; every challenge
  remains winnable at par.

## Docs

Update `docs/LESSONS.md` (Act II lesson list, move these items out of the roadmap) and the
README journey table.

## Out of scope

Counts with text objects (`2diw`), `a`-word objects (`daw`), other text objects (`i"`,
`ip`), and the rest of the Solid phase (`e`, `f`/`t`, `:%s`).
