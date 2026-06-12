package game

// sound.go — minimal, dependency-free terminal-bell sound effects.
//
// A TUI can't synthesize audio without CGO (which the release build disables),
// so we use the one piece of sound every terminal already knows: the ASCII BEL
// (0x07). It's emitted out-of-band through a tea.Cmd so it never lands inside an
// ANSI escape sequence or disturbs the alt-screen frame buffer. Distinct events
// ring a different number of bells; the master toggle is the M key.

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Sfx identifies a sound event. The terminal bell can't vary pitch, so we vary
// the *count* of bells to give events a recognisable rhythm.
type Sfx int

const (
	SfxHit   Sfx = iota // paddle catch — single, throttled
	SfxMiss             // ball lost — single
	SfxPower            // power-up collected — double
	SfxPhase            // difficulty up — double
	SfxStart            // countdown GO — single
	SfxOver             // game over — triple
	SfxBest             // new personal best — triple
)

// bells returns how many bell pulses an event rings.
func (s Sfx) bells() int {
	switch s {
	case SfxPower, SfxPhase:
		return 2
	case SfxOver, SfxBest:
		return 3
	default:
		return 1
	}
}

// requestSfx queues a sound for the next render flush, honouring the master
// toggle. SfxHit is rate-limited by hitBellCD so a fast rally produces a
// rhythmic tick rather than a continuous tone.
func (m *Model) requestSfx(s Sfx) {
	if !m.soundOn {
		return
	}
	if s == SfxHit {
		if m.hitBellCD > 0 {
			return
		}
		m.hitBellCD = HitBellGap
	}
	if n := s.bells(); n > m.bellCount {
		m.bellCount = n
	}
}

// flushBell drains any queued bells into a command and clears the queue.
// Returns nil when nothing is pending.
func (m *Model) flushBell() tea.Cmd {
	if m.bellCount == 0 {
		return nil
	}
	n := m.bellCount
	m.bellCount = 0
	return bellCmd(n)
}

// bellCmd writes n BEL bytes to the terminal, spaced so multi-bell events are
// audibly distinct. It runs in bubbletea's command goroutine, off the render
// loop, so the spacing never stalls the frame.
func bellCmd(n int) tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < n; i++ {
			_, _ = os.Stdout.Write([]byte{0x07})
			if i < n-1 {
				time.Sleep(85 * time.Millisecond)
			}
		}
		return nil
	}
}
