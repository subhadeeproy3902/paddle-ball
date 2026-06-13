package game

// update.go — bubbletea Update function.
// Key handling fix: paddle target is moved DIRECTLY in the key handler.
// harmonica spring (in physics tick) drives the actual paddle position.
// No held-key state that gets cleared every frame.

import (
	"math"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/subhadeeproy3902/paddle-ball/store"
	"github.com/subhadeeproy3902/paddle-ball/ui"
)

// Update is the bubbletea Update function — called for every message.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ── terminal resize ────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcPlayArea()
		m.clampPaddleTarget()
		return m, nil

	// ── progress bar animation frames (own chain, NOT a game tick) ─────────
	case progress.FrameMsg:
		bar, cmd := m.puBar.Update(msg)
		m.puBar = bar.(progress.Model)
		return m, cmd

	// ── keyboard ───────────────────────────────────────────────────────────
	case tea.KeyMsg:
		ks := msg.String()
		if ks == "ctrl+c" || (m.appPhase == PhaseTitle && (ks == "q" || ks == "Q")) {
			return m, tea.Quit
		}
		m = m.handleKey(msg)
		return m, nil

	// ── mouse: paddle follows a deliberate LEFT-drag while playing ─────────
	//   Gated to a held left button so a stray wheel scroll, hover, or
	//   click-release never yanks the paddle. Keyboard control stays precise.
	case tea.MouseMsg:
		if m.appPhase == PhasePlaying &&
			msg.Button == tea.MouseButtonLeft &&
			(msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionMotion) {
			m.setPaddleCenter(float64(msg.X))
		}
		return m, nil

	// ── game tick — the SINGLE source that re-arms the next tick ───────────
	case TickMsg:
		m = m.tick(time.Time(msg))
		return m, schedTick()
	}

	return m, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Key handling
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) handleKey(k tea.KeyMsg) Model {
	ks := k.String()

	switch m.appPhase {

	// ── TITLE ──────────────────────────────────────────────────────────────
	case PhaseTitle:
		switch ks {
		case "up", "k", "w":
			m.menuSel = (m.menuSel + 3) % 4
		case "down", "j":
			m.menuSel = (m.menuSel + 1) % 4
		case "enter", " ":
			m.mode = GameMode(m.menuSel)
			m.startCountdown()
		case "1":
			m.mode = ModeClassic
			m.startCountdown()
		case "2":
			m.mode = ModeArcade
			m.startCountdown()
		case "3":
			m.mode = ModeZen
			m.startCountdown()
		case "4":
			m.mode = ModeTimeTrial
			m.startCountdown()
		case "s", "S":
			m.appPhase = PhaseLeaderboard
			m.confirmClear = false
			recs, _ := m.st.LoadAll()
			m.scores = recs
		case "t", "T":
			m.cycleTheme()
		case "m", "M":
			m.toggleMute()
		case "?", "h", "H":
			m.appPhase = PhaseHelp
		}

	// ── COUNTDOWN ──────────────────────────────────────────────────────────
	case PhaseCountdown:
		switch ks {
		case "q", "Q", "esc":
			m.appPhase = PhaseTitle
		}

	// ── PLAYING ────────────────────────────────────────────────────────────
	// KEY FIX: We shift paddleTargX directly here. The spring in physics.go
	// smoothly chases the target every tick. No held-key state cleared.
	case PhasePlaying:
		switch ks {
		case "left", "a", "A", "h":
			m.movePaddleTarget(-KeyMoveStep)
		case "right", "d", "D", "l":
			m.movePaddleTarget(+KeyMoveStep)
		case "p", "P", " ":
			m.appPhase = PhasePaused
		case "t", "T":
			m.cycleTheme()
		case "m", "M":
			m.toggleMute()
		case "?":
			m.appPhase = PhaseHelp
		case "q", "Q", "esc":
			m.appPhase = PhaseTitle
		}

	// ── PAUSED ─────────────────────────────────────────────────────────────
	case PhasePaused:
		switch ks {
		case "p", "P", " ", "esc":
			m.appPhase = PhasePlaying
			m.lastTick = time.Now()
		case "r", "R":
			m.startCountdown()
		case "q", "Q":
			m.appPhase = PhaseTitle
		}

	// ── BALL LOST (modal) ──────────────────────────────────────────────────
	case PhaseBallLost:
		switch ks {
		case "enter", " ":
			m.resumeGame() // continue (choice) / skip the countdown
		case "m", "M":
			m.toggleMute()
		case "q", "Q", "esc":
			m.endGame() // give up — record the run and show the summary
		}

	// ── GAME OVER ──────────────────────────────────────────────────────────
	case PhaseGameOver:
		switch ks {
		case "r", "R", "enter", " ":
			m.startCountdown()
		case "s", "S":
			m.appPhase = PhaseLeaderboard
			m.confirmClear = false
			recs, _ := m.st.LoadAll()
			m.scores = recs
		case "m", "M":
			m.toggleMute()
		case "q", "Q", "esc":
			m.appPhase = PhaseTitle
		}

	// ── LEADERBOARD ────────────────────────────────────────────────────────
	case PhaseLeaderboard:
		// 'C' clears all history — first press arms, second press wipes.
		if ks == "c" || ks == "C" {
			if m.confirmClear {
				_ = m.st.Reset()
				m.confirmClear = false
				m.reloadScores()
				m.hiScore = m.st.HiScore(m.mode.Code())
			} else {
				m.confirmClear = true
			}
			return m
		}
		m.confirmClear = false // any other key cancels a pending clear
		switch ks {
		case "q", "Q", "esc", "enter":
			m.appPhase = PhaseTitle
		case "0":
			m.lbFilter = ""
			m.reloadScores()
		case "1":
			m.lbFilter = "classic"
			m.reloadScores()
		case "2":
			m.lbFilter = "arcade"
			m.reloadScores()
		case "3":
			m.lbFilter = "zen"
			m.reloadScores()
		case "4":
			m.lbFilter = "timed"
			m.reloadScores()
		}

	// ── HELP ───────────────────────────────────────────────────────────────
	case PhaseHelp:
		m.appPhase = PhaseTitle
	}

	return m
}

// movePaddleTarget shifts the paddle target by delta, clamping within play area.
func (m *Model) movePaddleTarget(delta float64) {
	m.paddleTargX += delta
	m.clampPaddleTarget()
	m.lastKeyTime = time.Now()
	m.paddleDir = math.Copysign(1, delta)
}

// setPaddleCenter aims the paddle so its centre sits under the given column
// (used by mouse control). The spring still smooths the actual motion.
func (m *Model) setPaddleCenter(col float64) {
	m.paddleTargX = col - float64(m.paddleW)/2
	m.clampPaddleTarget()
	m.lastKeyTime = time.Now()
}

// toggleMute flips the master sound switch and persists it.
func (m *Model) toggleMute() {
	m.soundOn = !m.soundOn
	m.st.SaveMuted(!m.soundOn)
}

func (m *Model) clampPaddleTarget() {
	max := float64(m.playW-1) - float64(m.paddleW)
	if m.paddleTargX < 0 {
		m.paddleTargX = 0
	}
	if m.paddleTargX > max {
		m.paddleTargX = max
	}
}

func (m *Model) cycleTheme() {
	m.themeIdx = (m.themeIdx + 1) % ui.ThemeCount
	m.st.SaveTheme(m.themeIdx)
	// Re-tint the power-up progress bar to the new palette.
	t := m.theme()
	m.puBar = progress.New(
		progress.WithGradient(t.Faint, t.Accent),
		progress.WithWidth(12),
		progress.WithoutPercentage(),
	)
}

func (m *Model) reloadScores() {
	all, _ := m.st.LoadAll()
	if m.lbFilter == "" {
		m.scores = all
		return
	}
	out := all[:0]
	for _, r := range all {
		if r.Mode == m.lbFilter {
			out = append(out, r)
		}
	}
	m.scores = out
}

// ─────────────────────────────────────────────────────────────────────────────
// Tick dispatcher
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) tick(now time.Time) Model {
	switch m.appPhase {
	case PhaseTitle:
		m.titleFrame++
		return m
	case PhaseCountdown:
		return m.tickCountdown(now)
	case PhasePlaying:
		return m.tickGame(now)
	case PhaseBallLost:
		return m.tickBallLost(now)
	}
	return m
}

// tickBallLost drives the auto-resume countdown. Choice modals (Zen) wait for a
// keypress and ignore the clock.
func (m Model) tickBallLost(now time.Time) Model {
	if m.lostChoice {
		m.lastTick = now
		return m
	}
	if m.lastTick.IsZero() {
		m.lastTick = now
		return m
	}
	dt := now.Sub(m.lastTick).Seconds()
	m.lastTick = now
	m.resumeTTL -= dt
	if m.resumeTTL <= 0 {
		m.resumeCount--
		m.resumeTTL = 0.85
		if m.resumeCount > 0 {
			m.requestSfx(SfxMenu)
		} else {
			m.resumeGame()
		}
	}
	return m
}

func (m Model) tickCountdown(now time.Time) Model {
	if m.lastTick.IsZero() {
		m.lastTick = now
		return m
	}
	dt := now.Sub(m.lastTick).Seconds()
	m.lastTick = now
	m.cdTTL -= dt
	if m.cdTTL <= 0 {
		if m.countdown > 0 {
			m.countdown--
			m.cdTTL = 1.0
			if m.countdown == 0 {
				m.requestSfx(SfxStart) // "GO!"
			}
		} else {
			m.startPlaying()
		}
	}
	return m
}

func (m Model) tickGame(now time.Time) Model {
	if m.lastTick.IsZero() {
		m.lastTick = now
		return m
	}
	dt := now.Sub(m.lastTick).Seconds()
	if dt > 0.05 {
		dt = 0.05
	} // cap: prevents physics explosion after alt-tab
	m.lastTick = now
	m.elapsed = now.Sub(m.gameStart)

	// Time Trial end
	if m.mode == ModeTimeTrial && m.elapsed >= m.timeLimit {
		m.endGame()
		return m
	}

	m.stepPaddleSpring(dt)
	m.updateBall(dt)
	m.updateParticles(dt)
	m.updateFloatTexts(dt)

	if m.mode == ModeArcade || m.mode == ModeZen {
		m.updateFallingPUs(dt)
	}
	m.stepActivePU(dt)

	if m.bannerTTL > 0 {
		m.bannerTTL -= dt
	}
	if m.paddleFlash > 0 {
		m.paddleFlash -= dt
		if m.paddleFlash < 0 {
			m.paddleFlash = 0
		}
	}
	return m
}

// ─────────────────────────────────────────────────────────────────────────────
// Spring step for paddle (harmonica)
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) stepPaddleSpring(dt float64) {
	prevX := m.paddleX
	// harmonica.Spring.Update advances the spring by one time-step (1/FPS).
	// Critically damped (SpringDamp = 1.0): the paddle glides to the target and
	// STOPS — no overshoot, no jelly-wobble after the key is released.
	m.paddleX, m.paddleVX = m.paddleSpring.Update(m.paddleX, m.paddleVX, m.paddleTargX)

	// Snap-and-settle: once we're within a fraction of a cell and barely moving,
	// pin to the target and zero the velocity. Kills any perpetual micro-tail or
	// sub-pixel jitter so the paddle is dead still when no key is pressed.
	if math.Abs(m.paddleX-m.paddleTargX) < 0.08 && math.Abs(m.paddleVX) < 0.5 {
		m.paddleX = m.paddleTargX
		m.paddleVX = 0
	}

	// Record actual velocity for spin transfer.
	if dt > 0 {
		m.paddleLastVX = (m.paddleX - prevX) / dt
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Float text tick
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) updateFloatTexts(dt float64) {
	alive := m.floatTxts[:0]
	for _, ft := range m.floatTxts {
		ft.Life -= ft.Decay * dt
		ft.Y -= 2.5 * dt // drift upward
		if ft.Life > 0 {
			alive = append(alive, ft)
		}
	}
	m.floatTxts = alive
}

// ─────────────────────────────────────────────────────────────────────────────
// Active power-up timer
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) stepActivePU(dt float64) {
	if m.activePU == nil || m.activePU.Total == 0 {
		return
	}
	m.activePU.TTL -= dt
	if m.activePU.TTL <= 0 {
		m.expirePU()
	}
}

func (m *Model) expirePU() {
	if m.activePU == nil {
		return
	}
	switch m.activePU.Kind {
	case PUWidePaddle:
		m.paddleW -= 3
		if m.paddleW < m.curPhase.PaddleW {
			m.paddleW = m.curPhase.PaddleW
		}
	}
	m.activePU = nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Falling power-ups
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) updateFallingPUs(dt float64) {
	alive := m.fallingPUs[:0]
	for _, pu := range m.fallingPUs {
		pu.Y += pu.VY * dt
		pu.AnimTTL -= dt
		if pu.AnimTTL <= 0 {
			pu.AnimStep = math.Mod(pu.AnimStep+1, 3)
			pu.AnimTTL = 0.2
		}

		// Caught by the paddle?
		px := int(pu.X + 0.5)
		py := int(pu.Y + 0.5)
		paddleRow := m.playH - PaddleRow
		if py >= paddleRow-1 && py <= paddleRow+1 &&
			px >= int(m.paddleX) && px < int(m.paddleX)+m.paddleW {
			m.activatePU(pu.Kind)
			continue
		}
		if pu.Y >= float64(m.playH) {
			continue // gone off-screen
		}
		alive = append(alive, pu)
	}
	m.fallingPUs = alive
}

func (m *Model) spawnPU() {
	kinds := []PowerUpKind{PUWidePaddle, PUSlowMo, PUFirePaddle, PUIronShield, PUGhost}
	kind := kinds[randN(len(kinds))]
	if randF() < 0.15 {
		kind = PUBomb
	}
	xMax := m.playW - 4
	if xMax < 2 {
		xMax = 2
	}
	x := float64(2 + randN(xMax-2))
	m.fallingPUs = append(m.fallingPUs, FallingPU{
		X: x, Y: 1.0, Kind: kind, VY: 6.0, AnimTTL: 0.2,
	})
}

func (m *Model) activatePU(kind PowerUpKind) {
	if m.activePU != nil {
		m.expirePU()
	}
	dur := kind.Duration()
	m.activePU = &ActivePU{Kind: kind, TTL: dur, Total: dur}

	switch kind {
	case PUWidePaddle:
		m.paddleW += 3
	case PUBomb:
		m.paddleW -= 2
		if m.paddleW < 3 {
			m.paddleW = 3
		}
		m.activePU.Total = 10
		m.activePU.TTL = 10
	case PUIronShield:
		m.shieldActive = true
		m.activePU.Total = 0
	case PUGhost:
		m.ghostActive = true
		m.activePU.Total = 0
	}

	m.floatTxts = append(m.floatTxts, FloatText{
		X:     m.paddleX + float64(m.paddleW)/2 - 4,
		Y:     float64(m.playH - PaddleRow - 2),
		Text:  kind.Name(),
		Color: kind.Color(),
		Life:  1.0, Decay: 0.7,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Resize helpers
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) recalcPlayArea() {
	// Header: 2 rows, footer: 2 rows, borders: 2 rows
	m.playH = m.height - 6
	if m.playH < 12 {
		m.playH = 12
	}
	// Left/right borders: 2 cols
	m.playW = m.width - 2
	if m.playW < 30 {
		m.playW = 30
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// End game
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) endGame() {
	m.appPhase = PhaseGameOver
	m.spawnExplosion()
	_ = m.st.Save(store.ScoreRecord{
		Mode:        m.mode.Code(),
		Score:       m.score,
		HighStreak:  m.maxStreak,
		BallsCaught: m.catches,
		BallsMissed: m.misses,
		MaxPhase:    m.curPhase.Num,
		DurationSec: int(m.elapsed.Seconds()),
		Timestamp:   time.Now(),
		Version:     "1.0.0",
	})
	if m.score > 0 && m.score >= m.hiScore {
		m.hiScore = m.score
		m.requestSfx(SfxBest)
	} else {
		m.requestSfx(SfxOver)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Small random helpers (avoids dot-importing math/rand everywhere)
// ─────────────────────────────────────────────────────────────────────────────

func randN(n int) int { return rand.Intn(n) }
func randF() float64  { return rand.Float64() }
