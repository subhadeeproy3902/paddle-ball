package game

import (
	"math"
	"math/rand"
)

// paddleRowY returns the Y coordinate of the paddle inside the play area.
func (m *Model) paddleRowY() int { return m.playH - PaddleRow }

// updateBall advances the ball one physics step and resolves all collisions.
// Vertical game: paddle at bottom (horizontal), ball bounces off L/R/Top walls.
func (m *Model) updateBall(dt float64) {
	speedMult := 1.0
	if m.activePU != nil && m.activePU.Kind == PUSlowMo {
		speedMult = 0.60
	}

	nx := m.ball.X + m.ball.VX*speedMult*dt
	ny := m.ball.Y + m.ball.VY*speedMult*dt

	// ── Left wall ─────────────────────────────────────────────────────────
	if nx <= 0 {
		nx = -nx
		m.ball.VX = math.Abs(m.ball.VX)
		m.spawnWallParticles(0, int(ny), "lr")
	}

	// ── Right wall ────────────────────────────────────────────────────────
	rEdge := float64(m.playW - 1)
	if nx >= rEdge {
		nx = 2*rEdge - nx
		m.ball.VX = -math.Abs(m.ball.VX)
		m.spawnWallParticles(m.playW-1, int(ny), "lr")
	}

	// ── Top wall ──────────────────────────────────────────────────────────
	if ny <= 0 {
		ny = -ny
		m.ball.VY = math.Abs(m.ball.VY)
		m.spawnWallParticles(int(nx), 0, "top")
	}

	// ── Paddle collision ──────────────────────────────────────────────────
	pRowY := float64(m.paddleRowY())
	// Detect the ball crossing the paddle row (moving downward, VY > 0)
	if m.ball.VY > 0 && m.ball.Y < pRowY && ny >= pRowY {
		paddleLeft := m.paddleX
		paddleRight := m.paddleX + float64(m.paddleW)

		if nx >= paddleLeft && nx <= paddleRight {
			// ── HIT ──────────────────────────────────────────────────────
			if m.ghostActive {
				// Ghost power-up: pass through once
				m.ghostActive = false
				if m.activePU != nil && m.activePU.Kind == PUGhost {
					m.activePU = nil
				}
			} else {
				ny = pRowY - (ny - pRowY) // reflect back up
				m.resolvePaddleHit(nx, &ny)
			}
		} else {
			// ── MISS ─────────────────────────────────────────────────────
			m.handleMiss()
			return
		}
	}

	// ── Safety: keep ball inside vertical bounds ──────────────────────────
	if ny >= float64(m.playH-1) {
		ny = float64(m.playH) - 2
		m.ball.VY = -math.Abs(m.ball.VY)
	}

	// ── Trail update ──────────────────────────────────────────────────────
	cur := Pt{X: int(math.Round(m.ball.X)), Y: int(math.Round(m.ball.Y))}
	m.ball.Trail = append([]Pt{cur}, m.ball.Trail...)
	if len(m.ball.Trail) > m.curPhase.TrailLen+1 {
		m.ball.Trail = m.ball.Trail[:m.curPhase.TrailLen+1]
	}

	m.ball.X = nx
	m.ball.Y = ny
}

// resolvePaddleHit computes the new velocity after the ball hits the paddle.
func (m *Model) resolvePaddleHit(bx float64, by *float64) {
	paddleCX := m.paddleX + float64(m.paddleW)/2
	// relHit: -1.0 (left edge) … 0 (centre) … +1.0 (right edge)
	relHit := (bx - paddleCX) / (float64(m.paddleW) / 2)
	relHit = math.Max(-1, math.Min(1, relHit))

	maxAngle := math.Pi / 3 // 60° max deflection
	angle := relHit * maxAngle

	speed := math.Sqrt(m.ball.VX*m.ball.VX + m.ball.VY*m.ball.VY)
	// Never let the ball slow down below phase base speed
	minSpeed := BaseSpeed * m.curPhase.SpeedMult
	if speed < minSpeed {
		speed = minSpeed
	}

	// New velocity: upward (VY negative) with horizontal spread from angle
	m.ball.VX = speed * math.Sin(angle)
	m.ball.VY = -speed * math.Cos(angle) // negative = upward

	// Spin transfer: paddle lateral velocity adds horizontal component
	m.ball.VX += m.paddleLastVX * 0.30
	// Clamp spin so the ball can't go too horizontal
	maxVX := speed * math.Sin(maxAngle)
	m.ball.VX = math.Max(-maxVX, math.Min(maxVX, m.ball.VX))

	// Edge hit detection (outer 12% of paddle)
	isEdge := math.Abs(relHit) > 0.88

	// Visual feedback
	m.paddleFlash = 0.12
	m.spawnPaddleParticles(int(bx), m.paddleRowY())

	// Score + phase check
	m.scoreHit(isEdge)
	m.checkPhaseTransition()

	// Power-up spawn (Arcade / Zen)
	if m.mode == ModeArcade || m.mode == ModeZen {
		m.catchesSinceLastPU++
		if m.catchesSinceLastPU >= PUCatchInterval {
			m.catchesSinceLastPU = 0
			m.spawnPU()
		}
	}
}

// handleMiss is called when the ball passes the paddle.
func (m *Model) handleMiss() {
	m.misses++
	m.streak = 0

	// Iron shield: one auto-save
	if m.shieldActive {
		m.shieldActive = false
		if m.activePU != nil && m.activePU.Kind == PUIronShield {
			m.activePU = nil
		}
		// Bounce ball back upward at current position
		m.ball.VY = -math.Abs(m.ball.VY)
		m.ball.Y = float64(m.paddleRowY()) - 1
		return
	}

	if m.mode == ModeZen {
		m.resetBallOnly()
		return
	}

	m.lives--
	if m.lives <= 0 {
		m.endGame()
	} else {
		m.resetBallOnly()
	}
}

// resetBallOnly repositions the ball without touching the paddle or score.
func (m *Model) resetBallOnly() {
	speed := BaseSpeed * m.curPhase.SpeedMult
	angle := (rand.Float64() - 0.5) * math.Pi / 2.5
	m.ball = Ball{
		X:  float64(m.playW) / 2,
		Y:  float64(m.playH) / 3,
		VX: speed * math.Sin(angle),
		VY: speed * math.Cos(angle),
	}
}

// spawnExplosion creates the game-over particle burst at the ball's last position.
func (m *Model) spawnExplosion() {
	glyphs := []rune{'✦', '✧', '★', '✶', '✷', '✸', '·', '*', '◆', '◇'}
	colors := []string{"#FFD700", "#FF5370", "#00FFFF", "#C3E88D", "#FFCB6B", "#89DDFF", "#FF8C00"}
	for i := 0; i < 18; i++ {
		a := rand.Float64() * 2 * math.Pi
		s := 6.0 + rand.Float64()*16.0
		m.particles = append(m.particles, Particle{
			X: m.ball.X, Y: m.ball.Y,
			VX:    s * math.Cos(a),
			VY:    s * math.Sin(a),
			Life:  1.0,
			Decay: 0.4 + rand.Float64()*0.5,
			Glyph: glyphs[rand.Intn(len(glyphs))],
			Color: colors[rand.Intn(len(colors))],
		})
	}
}