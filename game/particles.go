package game

import (
	"math"
	"math/rand"
)

// particles.go — a deliberately restrained particle system. Counts are small
// and colors come from the active theme (its muted/accent tones), so impacts
// read as a soft spark rather than confetti.

// updateParticles advances every live particle and removes dead ones.
func (m *Model) updateParticles(dt float64) {
	alive := m.particles[:0]
	for _, p := range m.particles {
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 4.0 * dt // gentle gravity
		p.Life -= p.Decay * dt
		if p.Life > 0 {
			alive = append(alive, p)
		}
	}
	m.particles = alive
}

// spawnWallParticles creates a small, quiet burst when the ball hits a wall.
func (m *Model) spawnWallParticles(x, y int) {
	col := m.theme().Wall
	glyphs := []rune{'·', '˙', '∘', '•'}
	n := 2 + rand.Intn(2)
	for i := 0; i < n; i++ {
		a := rand.Float64() * 2 * math.Pi
		s := 3.0 + rand.Float64()*6.0
		m.particles = append(m.particles, Particle{
			X: float64(x), Y: float64(y),
			VX:    s * math.Cos(a),
			VY:    s * math.Sin(a),
			Life:  1.0,
			Decay: 2.4 + rand.Float64()*2.0,
			Glyph: glyphs[rand.Intn(len(glyphs))],
			Color: col,
		})
	}
}

// spawnPaddleParticles creates the spark burst when the ball strikes the paddle.
// It fires an upward hemisphere in the theme accent + ball tones.
func (m *Model) spawnPaddleParticles(x, y int) {
	t := m.theme()
	glyphs := []rune{'✦', '✧', '·', '˙'}
	colors := []string{t.Accent, t.Ball, t.Muted}
	n := 3 + rand.Intn(3)
	for i := 0; i < n; i++ {
		a := rand.Float64()*math.Pi + math.Pi // upward hemisphere (π … 2π)
		s := 4.0 + rand.Float64()*10.0
		m.particles = append(m.particles, Particle{
			X: float64(x), Y: float64(y),
			VX:    s * math.Cos(a),
			VY:    s * math.Sin(a),
			Life:  1.0,
			Decay: 2.2 + rand.Float64()*2.0,
			Glyph: glyphs[rand.Intn(len(glyphs))],
			Color: colors[rand.Intn(len(colors))],
		})
	}
}
