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
	// Phase bonus
	switch m.curPhase.Num {
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

	// Rally milestone bonuses
	bonus := 0
	switch m.streak {
	case 50:
		bonus = 25
	case 100:
		bonus = 75
	case 200:
		bonus = 200
	}
	earned += bonus
	m.score += earned

	// Floating score label
	label := fmt.Sprintf("+%d", earned)
	if mult > 1.0 {
		label += fmt.Sprintf(" ×%.1g", mult)
	}
	m.floatTxts = append(m.floatTxts, FloatText{
		X:     m.paddleX + float64(m.paddleW)/2 - float64(len(label))/2,
		Y:     float64(m.paddleRowY() - 1),
		Text:  label,
		Color: scoreColor(earned),
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

// checkPhaseTransition upgrades difficulty when score crosses a threshold.
func (m *Model) checkPhaseTransition() {
	np := PhaseForScore(m.score)
	if np.Num > m.curPhase.Num {
		m.curPhase = np
		m.paddleW = np.PaddleW
		// Clamp paddle position after size change
		m.clampPaddleTarget()

		// Speed-bump the ball to the new phase speed
		speed := BaseSpeed * np.SpeedMult
		curr := math.Sqrt(m.ball.VX*m.ball.VX + m.ball.VY*m.ball.VY)
		if curr > 0 {
			m.ball.VX = m.ball.VX / curr * speed
			m.ball.VY = m.ball.VY / curr * speed
		}

		m.bannerText = fmt.Sprintf("  %s  PHASE %d — %s  %s  ",
			np.Emoji, np.Num, np.Name, np.Emoji)
		m.bannerColor = np.Color
		m.bannerTTL = 1.8
	}
}

// RankForScore returns a display rank and accent colour for a final score.
func RankForScore(score int) (string, string) {
	switch {
	case score >= 500:
		return "🏆 God Mode", "#FFD700"
	case score >= 200:
		return "💎 Grandmaster", "#89DDFF"
	case score >= 100:
		return "🌟 Legend", "#C3E88D"
	case score >= 50:
		return "⚡ Speedster", "#FFCB6B"
	case score >= 25:
		return "🔥 Baller", "#FF8C00"
	case score >= 10:
		return "🎮 Rookie", "#FF5370"
	default:
		return "🐣 Hatchling", "#AAAAAA"
	}
}

func scoreColor(pts int) string {
	switch {
	case pts >= 10:
		return "#FF5370"
	case pts >= 5:
		return "#FFCB6B"
	case pts >= 3:
		return "#C3E88D"
	default:
		return "#00FFFF"
	}
}