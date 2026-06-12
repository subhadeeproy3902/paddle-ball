//go:build shots

// Frame generator for the marketing screenshot — excluded from normal builds
// and CI. Regenerate with:  go test -tags shots ./game/ -run TestShot
// then:  python tools/shot.py game/zz_play.ansi "paddle-ball — arcade" tools/terminal.html

package game

import (
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Temporary: renders real game frames as truecolor ANSI to files, for building
// the marketing screenshot. Run: go test ./game/ -run TestShot -v
// Deleted before commit.
func TestShot(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)

	dump := func(path string, m Model) {
		if err := os.WriteFile(path, []byte(m.View()), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// ── Gameplay scene (Arcade, mid-rally) ──────────────────────────────────
	mp := NewModel("", "")
	mp.themeIdx = 0 // Claude
	mp.width, mp.height = 92, 26
	mp.recalcPlayArea()
	mp.appPhase = PhasePlaying
	mp.mode = ModeArcade
	mp.lives = 3
	mp.score = 128
	mp.hiScore = 540
	mp.streak = 17
	mp.maxStreak = 17
	mp.catches = 90
	mp.curPhase = Phases[3] // Blazing
	mp.resetAll()
	mp.curPhase = Phases[3]
	mp.paddleW = Phases[3].PaddleW
	mp.ball.X, mp.ball.Y = 52, 9
	mp.ball.Trail = []Pt{{52, 9}, {50, 8}, {48, 7}, {46, 6}, {44, 5}}
	mp.paddleX = 40
	mp.paddleTargX = 40
	mp.gameStart = time.Now()
	// an active power-up in the footer
	mp.activePU = &ActivePU{Kind: PUFirePaddle, TTL: 9, Total: 15}
	dump("zz_play.ansi", mp)

	// ── Title scene ─────────────────────────────────────────────────────────
	mt := NewModel("", "")
	mt.themeIdx = 0
	mt.width, mt.height = 92, 26
	mt.recalcPlayArea()
	mt.menuSel = 1
	mt.hiScore = 128
	dump("zz_title.ansi", mt)
}
