# paddle-ball

> A minimalist, physics-based paddleball game for the terminal — sub-stepped
> collisions that never glitch, a spring-driven paddle (keys or mouse), five
> restrained themes, four modes, and a persistent score history. One binary.

[![Go Version](https://img.shields.io/badge/Go-1.21+-555?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/subhadeeproy3902/paddle-ball?style=flat&color=cc785c)](https://github.com/subhadeeproy3902/paddle-ball/releases)
[![License](https://img.shields.io/badge/license-MIT-a09d96?style=flat)](LICENSE)

---

## Features

- **Sub-stepped physics** — the ball advances in collision-safe sub-steps, so it never tunnels through a wall or resets mid-rally, even at top speed.
- **Spring-driven paddle** — a harmonica spring gives weighty, fluid control. Use the keys, or just move your **mouse** and the paddle follows.
- **Five quiet themes** — Claude (default), Mono, Nord, Moss, Ember. Each leans on a single accent — no rainbow, no neon, no glow.
- **Four modes** — Classic, Arcade (lives + power-ups), Zen, and Time Trial.
- **Six power-ups** — Wide Paddle, Slow Mo, Fire Paddle, Iron Shield, Ghost Ball, Bomb (Arcade / Zen).
- **Five difficulty phases** — auto-escalating by score, Warm Up → Insane.
- **Sound** — restrained terminal-bell feedback on key moments; toggle with `M`.
- **Score history** — persistent JSON store with an in-game leaderboard, per-mode filters, and lifetime stats.
- **Single binary** — pure Go, zero runtime dependencies, one-command install.

---

## Install

### One-line (macOS / Linux)
```bash
curl -fsSL https://raw.githubusercontent.com/subhadeeproy3902/paddle-ball/main/install.sh | bash
```

### Homebrew (macOS / Linux)
```bash
brew tap subhadeeproy3902/paddle-ball
brew install paddle-ball
```

### Go install
```bash
go install github.com/subhadeeproy3902/paddle-ball@latest
```

### Docker (no install)
```bash
docker run --rm -it ghcr.io/subhadeeproy3902/paddle-ball:latest
```

Also available via apt/rpm/apk packages and Scoop — see the [releases page](https://github.com/subhadeeproy3902/paddle-ball/releases).

---

## Usage

```
paddle-ball                     # title screen
paddle-ball play                # jump straight in (Classic)
paddle-ball play --mode arcade  # Arcade mode with power-ups
paddle-ball play --mode zen     # endless rally
paddle-ball play --mode timed   # 60-second blitz
paddle-ball play --theme nord   # start on a chosen theme
paddle-ball scores              # leaderboard
paddle-ball scores --all        # full history
paddle-ball scores --json       # raw JSON
paddle-ball reset               # wipe saved scores
paddle-ball version             # version info
```

## Controls

| Key | Action |
|---|---|
| `←` `→` / `A` `D` | Move the paddle |
| mouse | Paddle follows the cursor |
| `P` / `Space` | Pause / resume |
| `T` | Cycle color theme |
| `M` | Toggle sound |
| `C` | Clear score history (on the leaderboard) |
| `R` | Restart (pause / game over) |
| `?` / `H` | Help |
| `Q` / `Ctrl+C` | Quit |

---

## Difficulty phases

| Phase | Score | Speed | Paddle |
|---|---|---|---|
| Warm Up | 0+ | 100% | 14 |
| Heating Up | 10+ | 122% | 12 |
| On Fire | 25+ | 150% | 10 |
| Blazing | 50+ | 185% | 8 |
| Insane | 100+ | 230% | 6 |

---

## Development

```bash
git clone https://github.com/subhadeeproy3902/paddle-ball
cd paddle-ball
go mod tidy
go run .            # run the game
go test ./...       # unit tests (physics regression suite)
go build ./...      # build
```

### Release
```bash
git tag v1.0.0
git push origin v1.0.0   # tags trigger GitHub Actions → GoReleaser
```
Every push and PR runs build / vet / test; only `v*` tags run a release.

---

## Project structure

```
paddle-ball/
├── main.go                 Entry point (version injection)
├── cmd/root.go             Cobra CLI commands
├── game/
│   ├── model.go            Types, Model struct, constructor
│   ├── update.go           Bubble Tea Update (input, mouse, ticks)
│   ├── view.go             Renderer (minimal dark HUD)
│   ├── physics.go          Sub-stepped collision + paddle response
│   ├── physics_test.go     Collision regression tests
│   ├── particles.go        Restrained particle system
│   ├── scoring.go          Points, streaks, phase transitions
│   └── sound.go            Terminal-bell sound effects
├── ui/theme.go             Five color themes + lipgloss helpers
├── store/store.go          Score + config persistence (atomic JSON)
├── index.html              Landing page
├── .goreleaser.yaml        Cross-compile + publish pipeline
└── .github/workflows       CI (build/vet/test) + tagged release
```

---

## Links

- **GitHub** — [github.com/subhadeeproy3902/paddle-ball](https://github.com/subhadeeproy3902/paddle-ball)
- **Twitter** — [@mvp_Subha](https://twitter.com/mvp_Subha)
- **LinkedIn** — [subhadeep3902](https://linkedin.com/in/subhadeep3902)

---

## License

MIT © [Subhadeep Roy](https://github.com/subhadeeproy3902)

*Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss) from the
[Charm](https://charm.sh) ecosystem.*
