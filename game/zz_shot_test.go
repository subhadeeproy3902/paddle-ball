//go:build shots

// Frame generator for previews/marketing — excluded from normal builds and CI.
// Regenerate: go test -tags shots ./game/ -run TestShot

package game

import (
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestShot(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	dump := func(path string, m Model) {
		if err := os.WriteFile(path, []byte(m.View()), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	base := func(w, h int) Model {
		m := NewModel("", "")
		m.themeIdx = 0
		m.width, m.height = w, h
		m.recalcPlayArea()
		return m
	}

	// Title (tall enough for the big gradient art)
	mt := base(96, 42)
	mt.menuSel = 1
	mt.hiScore = 540
	mt.titleFrame = 36
	dump("zz_title.ansi", mt)

	// Gameplay
	mp := base(92, 26)
	mp.appPhase = PhasePlaying
	mp.mode = ModeArcade
	mp.lives = 3
	mp.score = 128
	mp.hiScore = 540
	mp.streak = 17
	mp.maxStreak = 17
	mp.catches = 90
	mp.curPhase = Phases[3]
	mp.resetAll()
	mp.curPhase = Phases[3]
	mp.paddleW = Phases[3].PaddleW
	mp.ball.X, mp.ball.Y = 52, 9
	mp.ball.Trail = []Pt{{52, 9}, {50, 8}, {48, 7}, {46, 6}, {44, 5}}
	mp.paddleX = 40
	mp.gameStart = time.Now()
	mp.activePU = &ActivePU{Kind: PUFirePaddle, TTL: 9, Total: 15}
	dump("zz_play.ansi", mp)

	// Ball-lost: Arcade countdown over the live field
	ma := base(92, 30)
	ma.appPhase = PhaseBallLost
	ma.mode = ModeArcade
	ma.lives = 2
	ma.resumeCount = 2
	ma.resetAll()
	ma.lives = 2
	dump("zz_lost_count.ansi", ma)
}
