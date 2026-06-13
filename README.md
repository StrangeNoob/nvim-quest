# nvim-quest

Learn Neovim through an epic three-act terminal quest — no Neovim required.

Real Vim keystrokes, instantly. Press `d` then `w` and the word vanishes,
exactly like the real editor. Earn stars by solving rooms under par (golf!),
build combo streaks, guard your hearts, and beat the clock in boss fights.

## The journey

- **Act I · The Cursor Dojo** — modes, `hjkl`, `w`/`b`, `0` `$` `gg` `G`
- **Act II · The Motion Crypts** — `i`/`a`, `x` `dw` `dd`, `yy` `p`, `cw` `cc`
- **Act III · The Neon Grid** — `/` search and `n`, count prefixes (`4w`, `3dd`)

Each act ends in a timed boss fight.

## Play

```sh
go run .            # start the game
go run . stats      # progress summary
go run . reset      # wipe progress
```

Or install it: `go install .` then run `nvim-quest` from anywhere
(lessons are embedded — no working-directory requirement).

Progress saves to `~/.nvim-quest/progress.json`.

## Develop

```sh
go test ./...
```

Lessons are JSON files in `assets/lessons/` — add one and the content
integrity test validates it automatically.

Curriculum inspired by
[Learn-Vim-and-NeoVim](https://github.com/rcallaby/Learn-Vim-and-NeoVim).

## Roadmap

- **Solid:** `e`, `o`/`O`, `P`, `ciw`, `f`/`t`, `:%s/old/new/g`, undo lesson
- **Everything:** visual mode, visual block, marks, registers, macros
