package game

import (
	"math"
	"testing"
)

// newTestModel builds a minimal in-bounds Model suitable for physics tests.
// It deliberately avoids any store/disk dependency; tests use Time Trial mode
// (which re-serves on a miss instead of calling endGame) or guarantee a catch.
func newTestModel(mode GameMode) Model {
	m := Model{
		appPhase: PhasePlaying,
		mode:     mode,
		lives:    1,
		playW:    72,
		playH:    20,
		curPhase: Phases[0],
		paddleW:  14,
		soundOn:  false,
		themeIdx: 0,
	}
	// Centre a wide paddle.
	m.paddleX = float64(m.playW)/2 - float64(m.paddleW)/2
	m.paddleTargX = m.paddleX
	return m
}

func ballInBounds(m *Model) bool {
	const eps = 1e-6
	return m.ball.X >= -eps && m.ball.X <= float64(m.playW-1)+eps &&
		m.ball.Y >= -eps && m.ball.Y <= float64(m.playH-1)+eps
}

// TestBallNeverEscapes hammers the ball at extreme speeds for thousands of
// frames and asserts it always stays inside the play field. This is the
// anti-tunnelling guarantee that the sub-stepped collision provides.
func TestBallNeverEscapes(t *testing.T) {
	m := newTestModel(ModeTimeTrial)
	speeds := []struct{ vx, vy float64 }{
		{900, 40}, {-850, 120}, {30, 1200}, {-1500, -1500}, {600, -900},
	}
	for _, s := range speeds {
		m.ball = Ball{X: float64(m.playW) / 2, Y: float64(m.playH) / 3, VX: s.vx, VY: s.vy}
		for frame := 0; frame < 2000; frame++ {
			m.updateBall(1.0 / 60.0)
			if !ballInBounds(&m) {
				t.Fatalf("ball escaped at frame %d with v=(%g,%g): pos=(%g,%g) bounds=%dx%d",
					frame, s.vx, s.vy, m.ball.X, m.ball.Y, m.playW, m.playH)
			}
		}
	}
}

// TestSideWallReflects verifies a ball driven into a side wall is pinned to the
// boundary and has its horizontal velocity reversed (no escape, no NaN).
func TestSideWallReflects(t *testing.T) {
	m := newTestModel(ModeTimeTrial)
	// Heading hard into the left wall, only mildly downward.
	m.ball = Ball{X: 2, Y: 5, VX: -400, VY: 10}
	m.updateBall(1.0 / 60.0)
	if m.ball.X < 0 {
		t.Fatalf("ball passed through left wall: X=%g", m.ball.X)
	}
	if m.ball.VX <= 0 {
		t.Fatalf("left-wall hit did not reverse VX: VX=%g", m.ball.VX)
	}
}

// TestCornerHitIsNotAFalseMiss reproduces the reported glitch: a ball diving
// into the bottom corner used to be judged against a stale pre-bounce X and
// spuriously "reset from centre" mid-rally. With swept paddle detection it must
// reflect off the side wall and land on the paddle as a clean catch.
func TestCornerHitIsNotAFalseMiss(t *testing.T) {
	m := newTestModel(ModeClassic)
	// Paddle pinned to the left, covering the corner landing zone [0..14].
	m.paddleX = 0
	m.paddleTargX = 0

	// Just above the paddle row, diving down-left toward the corner.
	pRow := float64(m.paddleRowY())
	m.ball = Ball{X: 4, Y: pRow - 1, VX: -700, VY: 150}

	m.updateBall(1.0 / 30.0) // a fat frame, to force the corner interaction

	if m.misses != 0 {
		t.Fatalf("corner catch was counted as a miss (the old glitch): misses=%d", m.misses)
	}
	if m.appPhase != PhasePlaying {
		t.Fatalf("game ended on a corner catch: phase=%v", m.appPhase)
	}
	if m.ball.VY >= 0 {
		t.Fatalf("ball was not bounced upward off the paddle: VY=%g", m.ball.VY)
	}
	if m.catches != 1 {
		t.Fatalf("expected exactly one catch, got %d", m.catches)
	}
}

// TestCleanMissIsCounted ensures a genuine miss (paddle elsewhere) is still
// detected — the fix must not over-correct into never missing.
func TestCleanMissIsCounted(t *testing.T) {
	m := newTestModel(ModeTimeTrial) // Time Trial re-serves, no endGame/store needed
	m.paddleX = 0
	m.paddleTargX = 0 // paddle far left

	pRow := float64(m.paddleRowY())
	m.ball = Ball{X: float64(m.playW) - 6, Y: pRow - 1, VX: 0, VY: 300} // drop on the right

	m.updateBall(1.0 / 30.0)

	if m.misses != 1 {
		t.Fatalf("expected the drop on the right to be a miss, got misses=%d", m.misses)
	}
}

// TestPaddleHitKeepsSpeed checks the bounce preserves at least the phase's base
// speed (no death-spiral into a stalled, near-zero-velocity ball).
func TestPaddleHitKeepsSpeed(t *testing.T) {
	m := newTestModel(ModeClassic)
	pRow := float64(m.paddleRowY())
	// Just above the paddle, slow enough that the min-speed clamp must engage,
	// but moving down enough to actually cross the paddle plane this frame.
	m.ball = Ball{X: float64(m.playW) / 2, Y: pRow - 0.1, VX: 4, VY: 12}
	m.updateBall(1.0 / 30.0)
	if m.catches != 1 {
		t.Fatalf("test setup did not produce a paddle hit: catches=%d", m.catches)
	}
	speed := math.Hypot(m.ball.VX, m.ball.VY)
	min := BaseSpeed * m.curPhase.SpeedMult
	if speed < min-1e-6 {
		t.Fatalf("ball stalled below phase base speed: got %g want >= %g", speed, min)
	}
}
