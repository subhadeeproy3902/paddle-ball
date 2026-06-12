package game

import (
	"math"
	"math/rand"
)

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

// spawnWallParticles creates a small burst when the ball hits a wall.
func (m *Model) spawnWallParticles(x, y int, side string) {
	var col string
	switch side {
	case "top":
		col = "#4ECDC4"
	default: // lr
		col = "#4ECDC4"
	}
	glyphs := []rune{'·', '˙', '′', '*', '•', '∘'}
	n := 3 + rand.Intn(4)
	for i := 0; i < n; i++ {
		var vx, vy float64
		switch side {
		case "top":
			vx = (rand.Float64() - 0.5) * 12
			vy = rand.Float64() * 8
		case "lr":
			if x == 0 {
				vx = rand.Float64() * 10
			} else {
				vx = -rand.Float64() * 10
			}
			vy = (rand.Float64() - 0.5) * 10
		}
		m.particles = append(m.particles, Particle{
			X: float64(x), Y: float64(y),
			VX:    vx,
			VY:    vy,
			Life:  1.0,
			Decay: 1.8 + rand.Float64()*2.0,
			Glyph: glyphs[rand.Intn(len(glyphs))],
			Color: col,
		})
	}
}

// spawnPaddleParticles creates the spark burst when the ball hits the paddle.
func (m *Model) spawnPaddleParticles(x, y int) {
	glyphs := []rune{'✦', '✧', '·', '*', '◆', '˙'}
	colors := []string{"#00FFFF", "#FFFFFF", "#FFD700", "#C3E88D"}
	n := 5 + rand.Intn(4)
	for i := 0; i < n; i++ {
		a := rand.Float64()*math.Pi + math.Pi // upward hemisphere (π to 2π)
		s := 4.0 + rand.Float64()*12.0
		m.particles = append(m.particles, Particle{
			X: float64(x), Y: float64(y),
			VX:    s * math.Cos(a),
			VY:    s * math.Sin(a),
			Life:  1.0,
			Decay: 2.0 + rand.Float64()*2.0,
			Glyph: glyphs[rand.Intn(len(glyphs))],
			Color: colors[rand.Intn(len(colors))],
		})
	}
}