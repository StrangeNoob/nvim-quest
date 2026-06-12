# nvim-quest

`nvim-quest` is a beginner-friendly terminal game for learning Neovim and Vim basics. It teaches motions, editing, yank and paste, and search through small interactive missions with XP, streaks, hints, and badges.

V1 uses a lightweight Vim command simulator. It does not launch or require Neovim.

## Install

Requires Go 1.23 or newer.

```bash
go install .
```

Run from the project directory so the bundled `lessons/` directory is available.

## Run

```bash
go run . start
go run . lesson 002-motions
go run . practice delete
go run . stats
go run . reset
```

Practice topics are `motions`, `insert`, `delete`, `yank-paste`, and `search`.

## Example Gameplay

```text
nvim-quest  Lesson 002-motions
Reach the target

Objective
Move the cursor to the word TARGET using h, j, k, l.

Buffer
 1  alpha beta gamma
 2  delta TARGET omega

Command
> j

Feedback
Command applied. Keep going.
```

Type a command and press Enter. Use `?` to toggle the hint, `q` to quit, and `Ctrl+c` to quit at any time.

Supported commands:

- Motions: `h`, `j`, `k`, `l`, `w`, `b`, `0`, `$`, `gg`, `G`
- Editing: `x`, `dw`, `dd`, `yy`, `p`, `iTEXT<Esc>`, `aTEXT<Esc>`, `u`
- Search: `/word`, `n`

Progress is saved to `~/.nvim-quest/progress.json`.

## Lesson Roadmap

1. Meet Normal and Insert mode
2. Basic movement
3. Word and file motions
4. Insert and append
5. Delete with intent
6. Yank and paste
7. Search the buffer

## Development

```bash
gofmt -w .
go test ./...
```

## Future Ideas

- Real Neovim integration
- Daily challenge
- Leaderboard
- Config file
- Custom lessons
- Macros
- Visual mode
- Splits and buffers
