package game

import (
	"math"
	"math/rand"
	"time"
)

// physics.go — sub-stepped continuous collision.
//
// The ball moves at most ~1 cell per sub-step, so it can never tunnel through a
// wall or skip the paddle plane at high speed. Walls clamp-and-reflect (the ball
// is pinned exactly to the boundary, then its velocity flips) which is rock
// stable — no "escaped past the wall" states. The paddle plane is a swept test:
// we interpolate the ball's X at the exact row crossing, so a ball grazing a
// bottom corner reflects off the side wall first and is judged against the
// paddle at the real crossing point, not a stale pre-bounce X. That corner case
// was the source of the "ball resets from centre for no reason" glitch.

// paddleRowY returns the Y coordinate of the paddle inside the play area.
func (m *Model) paddleRowY() int { return m.playH - PaddleRow }

// updateBall advances the ball one frame, split into collision-safe sub-steps.
func (m *Model) updateBall(dt float64) {
	if m.hitBellCD > 0 {
		m.hitBellCD -= dt
	}
	if m.bounceCD > 0 {
		m.bounceCD -= dt
	}

	speedMult := 1.0
	if m.activePU != nil && m.activePU.Kind == PUSlowMo {
		speedMult = 0.60
	}

	// Pick a sub-step count so the ball travels ≤ ~0.9 cells per step.
	dist := math.Hypot(m.ball.VX, m.ball.VY) * speedMult * dt
	steps := int(math.Ceil(dist / 0.9))
	if steps < 1 {
		steps = 1
	}
	if steps > MaxSubSteps {
		steps = MaxSubSteps
	}
	sdt := dt / float64(steps)

	for i := 0; i < steps; i++ {
		if m.advanceBall(sdt, speedMult) {
			return // ball lost — miss already resolved, stop stepping
		}
	}
}

// advanceBall integrates one sub-step and resolves every collision for it.
// It returns true when the ball was lost (a miss was handled), signalling the
// caller to stop sub-stepping this frame.
func (m *Model) advanceBall(dt, speedMult float64) (lost bool) {
	nx := m.ball.X + m.ball.VX*speedMult*dt
	ny := m.ball.Y + m.ball.VY*speedMult*dt

	// ── Side walls: clamp to the boundary, flip horizontal velocity ─────────
	if nx <= 0 {
		nx = 0
		if m.ball.VX < 0 {
			m.ball.VX = -m.ball.VX
			m.requestSfx(SfxBounce)
		}
		m.spawnWallParticles(0, clampInt(int(math.Round(ny)), 0, m.playH-1))
	} else if rEdge := float64(m.playW - 1); nx >= rEdge {
		nx = rEdge
		if m.ball.VX > 0 {
			m.ball.VX = -m.ball.VX
			m.requestSfx(SfxBounce)
		}
		m.spawnWallParticles(m.playW-1, clampInt(int(math.Round(ny)), 0, m.playH-1))
	}

	// ── Top wall ────────────────────────────────────────────────────────────
	if ny <= 0 {
		ny = 0
		if m.ball.VY < 0 {
			m.ball.VY = -m.ball.VY
			m.requestSfx(SfxBounce)
		}
		m.spawnWallParticles(clampInt(int(math.Round(nx)), 0, m.playW-1), 0)
	}

	// ── Paddle plane (swept) ────────────────────────────────────────────────
	pRowY := float64(m.paddleRowY())
	if m.ball.VY > 0 && m.ball.Y <= pRowY && ny >= pRowY {
		// Interpolate the X at which the ball crosses the paddle row.
		hitX := nx
		if span := ny - m.ball.Y; span > 1e-9 {
			tHit := (pRowY - m.ball.Y) / span
			hitX = m.ball.X + (nx-m.ball.X)*tHit
		}

		if hitX >= m.paddleX && hitX <= m.paddleX+float64(m.paddleW) {
			if m.ghostActive {
				// Ghost ball: phase straight through, once.
				m.ghostActive = false
				if m.activePU != nil && m.activePU.Kind == PUGhost {
					m.activePU = nil
				}
			} else {
				ny = pRowY - (ny - pRowY) // reflect upward
				nx = hitX                 // settle at the contact point
				m.resolvePaddleHit(hitX, &ny)
			}
		} else {
			m.handleMiss()
			return true
		}
	}

	// ── Floor safety net (only reachable via Ghost pass-through) ────────────
	if ny >= float64(m.playH-1) {
		ny = float64(m.playH) - 2
		m.ball.VY = -math.Abs(m.ball.VY)
	}

	m.commitBallPos(nx, ny)
	return false
}

// commitBallPos writes the new ball position and extends the trail, sampling per
// sub-step so the trail stays gap-free even at high speed.
func (m *Model) commitBallPos(nx, ny float64) {
	cur := Pt{X: int(math.Round(nx)), Y: int(math.Round(ny))}
	if len(m.ball.Trail) == 0 || m.ball.Trail[0] != cur {
		m.ball.Trail = append([]Pt{cur}, m.ball.Trail...)
		maxLen := m.curPhase.TrailLen + 1
		if len(m.ball.Trail) > maxLen {
			m.ball.Trail = m.ball.Trail[:maxLen]
		}
	}
	m.ball.X = nx
	m.ball.Y = ny
}

// resolvePaddleHit computes the new velocity after the ball strikes the paddle.
func (m *Model) resolvePaddleHit(bx float64, by *float64) {
	paddleCX := m.paddleX + float64(m.paddleW)/2
	// relHit: -1 (left edge) … 0 (centre) … +1 (right edge)
	relHit := (bx - paddleCX) / (float64(m.paddleW) / 2)
	relHit = math.Max(-1, math.Min(1, relHit))

	maxAngle := math.Pi / 3 // 60° max deflection
	angle := relHit * maxAngle

	speed := math.Hypot(m.ball.VX, m.ball.VY)
	if minSpeed := BaseSpeed * m.curPhase.SpeedMult; speed < minSpeed {
		speed = minSpeed
	}

	// New velocity: upward (VY negative) with horizontal spread from the angle.
	m.ball.VX = speed * math.Sin(angle)
	m.ball.VY = -speed * math.Cos(angle)

	// Spin transfer: the paddle's lateral velocity nudges the ball sideways.
	m.ball.VX += m.paddleLastVX * 0.28
	maxVX := speed * math.Sin(maxAngle)
	m.ball.VX = math.Max(-maxVX, math.Min(maxVX, m.ball.VX))

	isEdge := math.Abs(relHit) > 0.86

	// Feedback
	m.paddleFlash = 0.12
	m.spawnPaddleParticles(int(math.Round(bx)), m.paddleRowY())
	m.requestSfx(SfxHit)

	// Score + difficulty
	m.scoreHit(isEdge)
	m.checkPhaseTransition()

	// Power-up cadence (Arcade)
	if m.mode == ModeArcade {
		m.catchesSinceLastPU++
		if m.catchesSinceLastPU >= PUCatchInterval {
			m.catchesSinceLastPU = 0
			m.spawnPU()
		}
	}
}

// handleMiss is called when the ball passes the paddle plane outside its span.
func (m *Model) handleMiss() {
	m.misses++
	m.streak = 0

	// Iron shield: one automatic save.
	if m.shieldActive {
		m.shieldActive = false
		if m.activePU != nil && m.activePU.Kind == PUIronShield {
			m.activePU = nil
		}
		m.ball.VY = -math.Abs(m.ball.VY)
		m.ball.Y = float64(m.paddleRowY()) - 1
		m.requestSfx(SfxPower)
		return
	}

	m.requestSfx(SfxMiss)

	switch m.mode {
	case ModeTimeTrial:
		// Blitz — brief, visible resume so it never just teleports to centre.
		m.enterBallLost("Ball out", 2)
	default: // Classic, Arcade
		m.lives--
		if m.lives <= 0 {
			m.endGame()
		} else {
			m.enterBallLost("You lost a ball", 3)
		}
	}
}

// enterBallLost pauses the ball and starts the auto-resume countdown. The play
// field stays live (paddle controllable) until the ball is re-served after a
// count of `count`.
func (m *Model) enterBallLost(msg string, count int) {
	m.appPhase = PhaseBallLost
	m.lostMsg = msg
	m.resumeCount = count
	m.resumeTTL = 0.85
	m.ball.Trail = nil
}

// resumeGame serves a fresh ball and returns to play.
func (m *Model) resumeGame() {
	m.appPhase = PhasePlaying
	m.lastTick = time.Now()
	m.resetBallOnly()
	m.requestSfx(SfxStart)
}

// resetBallOnly re-serves the ball from the top WITHOUT touching paddle or
// score. It serves directly above the paddle's current position with a gentle
// angle, so the resumed rally is always reachable — never an instant loss.
func (m *Model) resetBallOnly() {
	speed := BaseSpeed * m.curPhase.SpeedMult
	angle := (rand.Float64() - 0.5) * math.Pi / 6 // ±30°, kept narrow so it stays near the paddle
	m.ball = Ball{
		X:  m.paddleX + float64(m.paddleW)/2,
		Y:  float64(m.playH) / 3,
		VX: speed * math.Sin(angle),
		VY: speed * math.Cos(angle),
	}
}

// spawnExplosion creates the game-over burst at the ball's last position, drawn
// from the active theme so it never breaks the palette.
func (m *Model) spawnExplosion() {
	t := m.theme()
	glyphs := []rune{'✦', '✧', '·', '*', '◆', '◇', '•'}
	colors := []string{t.Accent, t.Ball, t.Muted, t.Trail[3], t.Trail[2]}
	for i := 0; i < 16; i++ {
		a := rand.Float64() * 2 * math.Pi
		s := 6.0 + rand.Float64()*15.0
		m.particles = append(m.particles, Particle{
			X: m.ball.X, Y: m.ball.Y,
			VX:    s * math.Cos(a),
			VY:    s * math.Sin(a),
			Life:  1.0,
			Decay: 0.45 + rand.Float64()*0.5,
			Glyph: glyphs[rand.Intn(len(glyphs))],
			Color: colors[rand.Intn(len(colors))],
		})
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
