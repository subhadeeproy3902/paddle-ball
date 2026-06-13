<div align="center">

<a href="https://pongball.mvp-subha.me">
  <img src="assets/banner.svg" alt="pong-ball — a minimalist, physics-based paddle game for your terminal" width="100%" />
</a>

<br />

[![Release](https://img.shields.io/github/v/release/subhadeeproy3902/pong-ball?style=flat-square&color=cc785c&labelColor=1f1e1b)](https://github.com/subhadeeproy3902/pong-ball/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/subhadeeproy3902/pong-ball/release.yml?style=flat-square&label=ci&color=5db8a6&labelColor=1f1e1b)](https://github.com/subhadeeproy3902/pong-ball/actions)
[![Go](https://img.shields.io/badge/Go-1.21+-a09d96?style=flat-square&logo=go&logoColor=white&labelColor=1f1e1b)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-a09d96?style=flat-square&labelColor=1f1e1b)](LICENSE)

**[Website](https://pongball.mvp-subha.me)** · **[Install](#install)** · **[Releases](https://github.com/subhadeeproy3902/pong-ball/releases)** · **[Controls](#controls)**

</div>

---

## Features

- **Sub-stepped physics** — the ball advances in collision-safe sub-steps, so it never tunnels through a wall or resets mid-rally, even at top speed.
- **Spring-driven paddle** — a harmonica spring gives weighty, fluid control. Use the keys, or just move your **mouse** and the paddle follows.
- **Five quiet themes** — Claude (default), Mono, Nord, Moss, Ember. Each leans on a single accent — no rainbow, no neon, no glow.
- **Three modes** — Classic, Arcade (lives + power-ups), and Time Trial.
- **Six power-ups** — Wide Paddle, Slow Mo, Fire Paddle, Iron Shield, Ghost Ball, Bomb (Arcade).
- **Five difficulty phases** — auto-escalating by score, Warm Up → Insane.
- **Real sound** — crisp arcade SFX from [soundcn](https://soundcn.xyz), embedded in the binary and played CGO-free (MCI on Windows, native players elsewhere); toggle with `M`.
- **Score history** — persistent JSON store with an in-game leaderboard, per-mode filters, and lifetime stats.
- **Single binary** — pure Go, zero runtime dependencies, one-command install.

---

## Gameplay

<div align="center">
  <img src="assets/screenshot.png" alt="pong-ball running in a terminal — Arcade mode with score, lives, streak, the ball and its trail, the coral paddle, and an active Fire Paddle power-up" width="820" />
  <br />
  <sub>Arcade · Claude theme — an actual rendered frame, not a mockup.</sub>
</div>

---

### macOS / Linux
```bash
# one-line installer (POSIX sh — no bash needed)
curl -fsSL https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/install.sh | sh

# Homebrew — tap this repo, then install (the formula lives in its Formula/ folder).
# `brew install <raw-url>` is NOT valid syntax; tap first.
brew tap subhadeeproy3902/pong-ball https://github.com/subhadeeproy3902/pong-ball
brew install subhadeeproy3902/pong-ball/pong-ball
```

### Windows
```powershell
# PowerShell one-liner
irm https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/install.ps1 | iex

# Scoop (manifest is self-hosted in this repo's bucket/ folder)
scoop install https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/bucket/pong-ball.json

# WinGet — manifests live in this repo's manifests/ folder (use the version dir)
winget install --manifest manifests/s/subhadeeproy3902/pong-ball/1.0.0
```

### Any platform
```bash
go install github.com/subhadeeproy3902/pong-ball@latest
docker run --rm -it ghcr.io/subhadeeproy3902/pong-ball:latest
```

Prebuilt binaries and `.deb`/`.rpm`/`.apk` packages are on the
[releases page](https://github.com/subhadeeproy3902/pong-ball/releases).
All package manifests (Homebrew/Scoop/WinGet) are kept **in this repo** —
`Formula/`, `bucket/`, `manifests/` — and refreshed on every release.

### Uninstall

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/uninstall.sh | sh
```
```powershell
# Windows
irm https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/uninstall.ps1 | iex
```

Both prompt for a **Y/N** confirmation, then remove the binary (every copy on
your `PATH`), its `PATH` entry, the saved scores/config, and the cached sound
files. (Scoop/Homebrew installs: use `scoop uninstall pong-ball` /
`brew uninstall pong-ball`.)

---

## Usage

```
pong-ball                     # title screen
pong-ball play                # jump straight in (Classic)
pong-ball play --mode arcade  # Arcade mode with power-ups
pong-ball play --mode timed   # 60-second blitz
pong-ball play --theme nord   # start on a chosen theme
pong-ball scores              # leaderboard
pong-ball scores --all        # full history
pong-ball scores --json       # raw JSON
pong-ball reset               # wipe saved scores
pong-ball version             # version info
```

## Controls

| Key | Action |
|---|---|
| `←` `→` / `A` `D` | Move the paddle |
| drag mouse | Hold the left button to move the paddle |
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
git clone https://github.com/subhadeeproy3902/pong-ball
cd pong-ball
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
pong-ball/
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
│   ├── sound.go            Sound-effect events + throttling
│   └── audio*.go            Embedded MP3s + CGO-free playback backends
├── ui/theme.go             Five color themes + lipgloss helpers
├── store/store.go          Score + config persistence (atomic JSON)
├── assets/                 Logo, banner, OG image, screenshot, icons
├── index.html              Landing page (pongball.mvp-subha.me)
├── .goreleaser.yaml        Cross-compile + publish pipeline
└── .github/workflows       CI (build/vet/test) + tagged release
```

---

## Links

- **GitHub** — [github.com/subhadeeproy3902/pong-ball](https://github.com/subhadeeproy3902/pong-ball)
- **Twitter** — [@mvp_Subha](https://twitter.com/mvp_Subha)
- **LinkedIn** — [subhadeep3902](https://linkedin.com/in/subhadeep3902)

---

## License

MIT © [Subhadeep Roy](https://github.com/subhadeeproy3902)

*Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss) from the
[Charm](https://charm.sh) ecosystem.*
