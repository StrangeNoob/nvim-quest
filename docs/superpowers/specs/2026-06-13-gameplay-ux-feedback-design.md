# nvim-quest — Gameplay UX Feedback Design

**Date:** 2026-06-13
**Status:** Approved for implementation

Three player-feedback improvements, all contained in `internal/ui`. No engine or
content changes.

## 1. Next-act announcement on boss clear

When an act's boss is cleared, the act-complete results screen names what was just
unlocked so progression is unmistakable:

```
⚔ ACT I COMPLETE ⚔

Sensei's Trial defeated!
+200 XP · 🏅 Dojo Graduate

──────────────────────────────────────
UNLOCKED ▶ ACT II · THE MOTION CRYPTS
learn: i/a · x · dw · dd · yy/p · cw · cc

[enter] enter the crypts
```

- The boss name comes from `lessons[lessonIdx].Boss.Name`.
- The next act and its teaches-summary come from `lessons[lessonIdx+1]`. A small
  `actSummary(act int) string` helper in `styles.go` maps act → the commands it teaches
  (1: motions; 2: edits; 3: search & counts).
- Only applies when there IS a next act. The final boss routes to #3 instead.

## 2. Why a heart was lost

On an invalid keystroke, the room shows a transient red line explaining the loss, e.g.:

```
✗ 'z' won't work here — that key isn't part of this room (lost a heart ♥)
```

- New `Model.heartMsg string`, set in `updateRoom` on `engine.EvInvalid` (before the
  heart decrement / reset logic), cleared at the start of any keystroke that is not
  invalid (so it disappears on the next valid move).
- Rendered in `viewRoom` with `dangerStyle`, between the HUD and the hint, only when
  non-empty and not flashing.
- The key is shown human-readably (`space` for `" "`); the message states it cost a
  heart, reinforcing the cause/effect the player asked about.

## 3. Final celebration

Clearing the final boss (the last lesson overall, Act III · The Grid Core) shows a
celebration instead of the standard act-complete screen:

```
        🎉  YOU MASTERED THE BLADE CURSOR  🎉

   The Grid Core is broken. The journey is complete.

   level 7 · 1840 XP · ⭐ 58/66 stars
   33 rooms cleared · 🏅 7 badges

           [enter] return to your journey
```

- New `Model.resGameComplete bool`, set true in `awardBoss` when
  `lessonIdx == len(lessons)-1`.
- `viewResults` checks `resGameComplete` first → celebration branch (totals from
  `prog`: level, XP, summed stars vs. max, completed count, badge count). `[enter]`
  returns to the world map (now fully ✓).
- Reuses the existing results navigation; no new screen enum value.

## Testing (`internal/ui/app_test.go`)

- **Next-act announcement:** drive a non-final boss to completion (e.g. Act I) and
  assert the results view contains the next act's name and "UNLOCKED".
- **Final celebration:** complete the final boss (Act III) and assert the view contains
  "MASTERED" and run totals, and `resGameComplete` is set.
- **Heart message:** an invalid key sets `heartMsg` and the room view shows it; a valid
  key clears `heartMsg`.

## Out of scope

A pre-boss "BOSS AHEAD" confirmation screen (considered, deferred), animations beyond the
existing flash, and any content/engine changes.
