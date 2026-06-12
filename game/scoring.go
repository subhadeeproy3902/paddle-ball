package game

import (
	"fmt"
	"math"
)

// scoreHit awards points for a successful paddle hit.
func (m *Model) scoreHit(edgeHit bool) {
	m.catches++
	m.streak++
	if m.streak > m.maxStreak {
		m.maxStreak = m.streak
	}

	pts := 1
	if edgeHit {
		pts += 2 // edge bonus
	}
	switch m.curPhase.Num { // phase bonus
	case 4:
		pts += 2
	case 5:
		pts += 4
	}

	mult := m.streakMult()
	if m.activePU != nil && m.activePU.Kind == PUFirePaddle {
		mult *= 2.0 // fire paddle: double score
	}

	earned := int(math.Round(float64(pts) * mult))

	// Rally milestone bonuses.
	bonus := 0
	switch m.streak {
	case 50:
		bonus = 25
	case 100:
		bonus = 75
	case 200:
		bonus = 200
	}
	if bonus > 0 {
		m.requestSfx(SfxPhase)
	}
	earned += bonus
	m.score += earned

	// Floating "+N" label — accent when boosted, muted otherwise.
	t := m.theme()
	col := t.Muted
	if mult > 1.0 || edgeHit || bonus > 0 {
		col = t.Accent
	}
	label := fmt.Sprintf("+%d", earned)
	if mult > 1.0 {
		label += fmt.Sprintf(" ×%.1g", mult)
	}
	m.floatTxts = append(m.floatTxts, FloatText{
		X:     m.paddleX + float64(m.paddleW)/2 - float64(len(label))/2,
		Y:     float64(m.paddleRowY() - 1),
		Text:  label,
		Color: col,
		Life:  1.0,
		Decay: 1.3,
	})
}

// streakMult returns the current score multiplier based on streak length.
func (m *Model) streakMult() float64 {
	switch {
	case m.streak >= 35:
		return 3.0
	case m.streak >= 20:
		return 2.0
	case m.streak >= 10:
		return 1.5
	default:
		return 1.0
	}
}

// checkPhaseTransition upgrades difficulty when the score crosses a threshold.
func (m *Model) checkPhaseTransition() {
	np := PhaseForScore(m.score)
	if np.Num <= m.curPhase.Num {
		return
	}
	m.curPhase = np
	m.paddleW = np.PaddleW
	m.clampPaddleTarget()

	// Re-normalise the ball to the new phase speed.
	speed := BaseSpeed * np.SpeedMult
	if curr := math.Hypot(m.ball.VX, m.ball.VY); curr > 0 {
		m.ball.VX = m.ball.VX / curr * speed
		m.ball.VY = m.ball.VY / curr * speed
	}

	m.bannerText = fmt.Sprintf("PHASE %d — %s", np.Num, np.Name)
	m.bannerColor = m.theme().Phase[np.Num-1]
	m.bannerTTL = 1.6
	m.requestSfx(SfxPhase)
}

// RankForScore returns a display rank and a tier index (0 = lowest). The view
// maps the tier onto the active theme's ramp so the rank colour stays on-palette.
func RankForScore(score int) (string, int) {
	switch {
	case score >= 500:
		return "God Mode", 6
	case score >= 200:
		return "Grandmaster", 5
	case score >= 100:
		return "Legend", 4
	case score >= 50:
		return "Speedster", 3
	case score >= 25:
		return "Baller", 2
	case score >= 10:
		return "Rookie", 1
	default:
		return "Hatchling", 0
	}
}
