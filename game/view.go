package game

// view.go — all rendering.
// Uses: lipgloss (styling), termenv (colour profile), bubbles/progress (PU bar).

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/subhadeeproy3902/paddle-ball/ui"
)

// colProfile is detected once at startup for fallback decisions.
var colProfile = termenv.ColorProfile()

// cell is one character slot in the play-area grid.
type cell struct {
	r     rune
	color string
	bold  bool
}

// ─────────────────────────────────────────────────────────────────────────────
// View — dispatcher
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width < 60 || m.height < 20 {
		return m.viewTooSmall()
	}
	t := &ui.Themes[m.themeIdx]
	switch m.appPhase {
	case PhaseTitle:
		return m.viewTitle(t)
	case PhaseCountdown:
		return m.viewCountdown(t)
	case PhasePlaying:
		return m.viewPlaying(t)
	case PhasePaused:
		return m.viewPaused(t)
	case PhaseGameOver:
		return m.viewGameOver(t)
	case PhaseLeaderboard:
		return m.viewLeaderboard(t)
	case PhaseHelp:
		return m.viewHelp(t)
	}
	return ""
}

// ─────────────────────────────────────────────────────────────────────────────
// Too-small guard
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewTooSmall() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5370")).Bold(true).Render(
		fmt.Sprintf("\n  Terminal too small (min 60×20)\n  Current: %d×%d\n  Please resize.", m.width, m.height))
}

// ─────────────────────────────────────────────────────────────────────────────
// Title screen
// ─────────────────────────────────────────────────────────────────────────────

const asciiLogo = `
 ██████╗  █████╗ ██████╗ ██████╗ ██╗     ███████╗
 ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██║     ██╔════╝
 ██████╔╝███████║██║  ██║██║  ██║██║     █████╗
 ██╔═══╝ ██╔══██║██║  ██║██║  ██║██║     ██╔══╝
 ██║     ██║  ██║██████╔╝██████╔╝███████╗███████╗
 ╚═╝     ╚═╝  ╚═╝╚═════╝ ╚═════╝ ╚══════╝╚══════╝`

func (m Model) viewTitle(t *ui.Theme) string {
	logo := ui.SB(t.Paddle).Render(asciiLogo)

	modes := []struct{ label, desc string }{
		{"[1] Classic   ", "One life · pure score chase"},
		{"[2] Arcade    ", "3 lives · power-ups"},
		{"[3] Zen       ", "Infinite lives · just vibe"},
		{"[4] Time Trial", "60-second blitz"},
	}

	var menu strings.Builder
	for i, mo := range modes {
		num := ui.SB(t.ScoreText).Render(mo.label)
		desc := ui.S(t.DimText).Render(mo.desc)
		line := "  " + num + desc
		if i == m.menuSel {
			line = ui.SBG(t.Paddle, t.HeaderBg).Bold(true).
				Render("▶ " + mo.label + mo.desc)
		}
		menu.WriteString(line + "\n")
	}

	menuBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(0, 2).Width(50).
		Render(menu.String() + "\n" +
			ui.S(t.DimText).Render("  [↑↓] Select  [Enter] Start  [S] Scores  [T] Theme  [?] Help"))

	hi := ui.SB(t.HiText).Render(fmt.Sprintf("🏆 Best: %d", m.hiScore))
	themeLabel := ui.S(t.DimText).Render("Theme: " + t.Name)

	footer := hi + "   " + themeLabel

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, logo, "", menuBox, "", footer))
}

// ─────────────────────────────────────────────────────────────────────────────
// Countdown
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewCountdown(t *ui.Theme) string {
	label := "GO!"
	if m.countdown > 0 {
		label = fmt.Sprintf("%d", m.countdown)
	}
	big := ui.SB(t.Paddle).
		Width(10).
		Render(lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Width(10).Render(label))
	sub := ui.S(t.DimText).Render("Mode: " + m.mode.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, big, "", sub))
}

// ─────────────────────────────────────────────────────────────────────────────
// Playing screen
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewPlaying(t *ui.Theme) string {
	header := m.buildHeader(t)
	area := m.buildPlayArea(t)
	footer := m.buildFooter(t)

	out := header + "\n" + area + "\n" + footer

	// Phase-transition banner (overlays above play area)
	if m.bannerTTL > 0 {
		banner := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.bannerColor)).Bold(true).
			Width(m.width).Align(lipgloss.Center).
			Background(lipgloss.Color("#080818")).
			Render(m.bannerText)
		out = header + "\n" + banner + "\n" + area + "\n" + footer
	}
	return out
}

func (m Model) buildHeader(t *ui.Theme) string {
	title := ui.SB(t.Paddle).Render("🏓 PADDLEBALL")
	modeS := ui.S(t.DimText).Render(" · " + m.mode.String())
	score := ui.SB(t.ScoreText).Render(fmt.Sprintf("SCORE %03d", m.score))
	hi := ui.S(t.HiText).Render(fmt.Sprintf("🏆 %d", m.hiScore))

	left := title + modeS
	right := score + "   " + hi
	gap := m.width - visLen(left) - visLen(right) - 2
	if gap < 1 {
		gap = 1
	}
	row1 := " " + left + strings.Repeat(" ", gap) + right

	// Row 2: lives / streak / phase / timer
	var parts []string
	if m.mode == ModeArcade {
		hearts := strings.Repeat("❤️ ", m.lives)
		parts = append(parts, ui.S(t.LivesColor).Render(strings.TrimSpace(hearts)))
	}
	if m.streak >= 5 {
		parts = append(parts, ui.SB(t.StreakText).Render(fmt.Sprintf("🔥 ×%d", m.streak)))
	}
	parts = append(parts, ui.S(m.curPhase.Color).Render(m.curPhase.Emoji+" "+m.curPhase.Name))
	if m.mode == ModeTimeTrial {
		remain := m.timeLimit - m.elapsed
		if remain < 0 {
			remain = 0
		}
		secs := int(remain.Seconds())
		tc := t.ScoreText
		if secs <= 10 {
			tc = t.StreakText
		}
		parts = append(parts, ui.SB(tc).Render(fmt.Sprintf("⏱ %ds", secs)))
	}
	row2 := " " + strings.Join(parts, "   ")

	st := lipgloss.NewStyle().Background(lipgloss.Color(t.HeaderBg)).Width(m.width)
	return st.Render(row1) + "\n" + st.Render(row2)
}

func (m Model) buildFooter(t *ui.Theme) string {
	var puStr string
	if m.activePU != nil {
		if m.activePU.Total > 0 {
			pct := m.activePU.TTL / m.activePU.Total
			bar := m.puBar.ViewAs(pct)
			label := ui.S(m.activePU.Kind.Color()).Render(
				string(m.activePU.Kind.Glyph()) + " " + m.activePU.Kind.Name())
			puStr = label + " " + bar + fmt.Sprintf(" %.0fs  ", m.activePU.TTL)
		} else {
			puStr = ui.S(m.activePU.Kind.Color()).Render(
				string(m.activePU.Kind.Glyph())+" "+m.activePU.Kind.Name()+" ✓") + "  "
		}
	}
	ctrl := ui.S(t.DimText).Render("[←→ / AD] Move  [P] Pause  [T] Theme  [?] Help  [Q] Quit")
	st := lipgloss.NewStyle().Background(lipgloss.Color(t.FooterBg)).Width(m.width)
	return st.Render(" " + puStr + ctrl)
}

// ─────────────────────────────────────────────────────────────────────────────
// Play-area grid renderer
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) buildPlayArea(t *ui.Theme) string {
	W, H := m.playW, m.playH
	// Allocate 2-D cell grid
	grid := make([][]cell, H)
	for y := range grid {
		grid[y] = make([]cell, W)
		for x := range grid[y] {
			grid[y][x] = cell{r: ' '}
		}
	}

	set := func(x, y int, r rune, color string) {
		if x >= 0 && x < W && y >= 0 && y < H {
			grid[y][x] = cell{r: r, color: color}
		}
	}
	setBold := func(x, y int, r rune, color string) {
		if x >= 0 && x < W && y >= 0 && y < H {
			grid[y][x] = cell{r: r, color: color, bold: true}
		}
	}

	// ── Top wall ────────────────────────────────────────────────────────────
	for x := 0; x < W; x++ {
		set(x, 0, '─', t.WallTB)
	}
	// ── Left wall ───────────────────────────────────────────────────────────
	for y := 0; y < H; y++ {
		set(0, y, '│', t.WallLR)
	}
	// ── Right wall ──────────────────────────────────────────────────────────
	for y := 0; y < H; y++ {
		set(W-1, y, '│', t.WallLR)
	}
	// Corners
	set(0, 0, '┌', t.WallTB)
	set(W-1, 0, '┐', t.WallTB)

	// ── Ball trail ───────────────────────────────────────────────────────────
	trailGlyphs := []rune{'∙', '░', '▒', '▓'}
	for i, pt := range m.ball.Trail {
		if i >= len(trailGlyphs) {
			break
		}
		idx := len(t.Trail) - 1
		if i < len(t.Trail) {
			idx = i
		}
		// Shift trail idx: 0=farthest=darkest
		idx = len(trailGlyphs) - 1 - idx
		tidx := len(t.Trail) - 1 - i
		if tidx < 0 {
			tidx = 0
		}
		set(pt.X, pt.Y, trailGlyphs[i], t.Trail[tidx])
	}

	// ── Ball ─────────────────────────────────────────────────────────────────
	bx := int(math.Round(m.ball.X))
	by := int(math.Round(m.ball.Y))
	if m.appPhase == PhasePlaying {
		setBold(bx, by, '●', t.Ball)
	}

	// ── Paddle ───────────────────────────────────────────────────────────────
	pRow := m.paddleRowY()
	padColor := t.Paddle
	switch {
	case m.paddleFlash > 0:
		padColor = "#FFFFFF"
	case m.activePU != nil && m.activePU.Kind == PUFirePaddle:
		padColor = "#FF8C00"
	case m.shieldActive:
		padColor = "#4ECDC4"
	}

	px := int(math.Round(m.paddleX))
	// Paddle glyph: ═ for normal, use double lines for flash
	padGlyph := '═'
	if m.paddleFlash > 0 {
		padGlyph = '▬'
	}
	for i := 0; i < m.paddleW; i++ {
		g := padGlyph
		if i == 0 || i == m.paddleW-1 {
			g = '╪' // end caps
		}
		setBold(px+i, pRow, g, padColor)
	}

	// ── Particles ────────────────────────────────────────────────────────────
	for _, p := range m.particles {
		px2 := int(math.Round(p.X))
		py2 := int(math.Round(p.Y))
		set(px2, py2, p.Glyph, p.Color)
	}

	// ── Falling power-ups ───────────────────────────────────────────────────
	puAnims := []rune{'▿', '▾', '▽'}
	for _, pu := range m.fallingPUs {
		fpx := int(math.Round(pu.X))
		fpy := int(math.Round(pu.Y))
		frame := int(pu.AnimStep) % 3
		var g rune
		if fpy%2 == 0 {
			g = pu.Kind.Glyph()
		} else {
			g = puAnims[frame]
		}
		setBold(fpx, fpy, g, pu.Kind.Color())
	}

	// ── Floating score texts ─────────────────────────────────────────────────
	for _, ft := range m.floatTxts {
		fy := int(math.Round(ft.Y))
		fx := int(math.Round(ft.X))
		for i, ch := range ft.Text {
			set(fx+i, fy, ch, ft.Color)
		}
	}

	// ── Render grid to string ────────────────────────────────────────────────
	var sb strings.Builder
	for _, row := range grid {
		// Build each row by grouping contiguous same-colour runs
		i := 0
		for i < len(row) {
			c := row[i]
			// Collect run of same colour
			j := i + 1
			for j < len(row) && row[j].color == c.color && row[j].bold == c.bold {
				j++
			}
			segment := string(collectRunes(row[i:j]))
			if c.color != "" {
				sty := lipgloss.NewStyle().Foreground(lipgloss.Color(c.color))
				if c.bold {
					sty = sty.Bold(true)
				}
				sb.WriteString(sty.Render(segment))
			} else {
				sb.WriteString(segment)
			}
			i = j
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func collectRunes(cells []cell) []rune {
	out := make([]rune, len(cells))
	for i, c := range cells {
		out[i] = c.r
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// Pause overlay
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewPaused(t *ui.Theme) string {
	dur := m.elapsed.Round(time.Second)
	content := ui.SB(t.Paddle).Render("⏸  PAUSED") + "\n\n" +
		fmt.Sprintf("Score:    %d\n", m.score) +
		fmt.Sprintf("Streak:   ×%d\n", m.streak) +
		fmt.Sprintf("Elapsed:  %s\n", fmtDur(dur)) +
		"\n" +
		ui.S(t.DimText).Render("[Space / P]  Resume\n[R]          Restart\n[Q]          Quit")

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(t.Paddle)).
		Padding(1, 4).Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ─────────────────────────────────────────────────────────────────────────────
// Game Over
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewGameOver(t *ui.Theme) string {
	rank, rankColor := RankForScore(m.score)
	newBest := m.score > 0 && m.score >= m.hiScore

	row := func(label, val string) string {
		return fmt.Sprintf("  %-18s│  %s\n", ui.S(t.DimText).Render(label), val)
	}
	content :=
		ui.SB("#FF5370").Width(44).Align(lipgloss.Center).Render("GAME  OVER") + "\n" +
			strings.Repeat("─", 44) + "\n" +
			row("Final Score", ui.SB(t.ScoreText).Render(fmt.Sprintf("%d", m.score))) +
			row("Rank", ui.SB(rankColor).Render(rank)) +
			row("Best Streak", ui.S(t.StreakText).Render(fmt.Sprintf("×%d", m.maxStreak))) +
			row("Balls Caught", fmt.Sprintf("%d", m.catches)) +
			row("Misses", fmt.Sprintf("%d", m.misses)) +
			row("Max Phase", m.curPhase.Emoji+" "+m.curPhase.Name) +
			row("Played", fmtDur(m.elapsed.Round(time.Second))) +
			strings.Repeat("─", 44)

	if newBest {
		content += "\n" + ui.SB("#FFD700").Width(44).Align(lipgloss.Center).
			Render("★  NEW PERSONAL BEST!  ★")
	}
	content += "\n" + ui.S(t.DimText).Render(
		"\n  [R / Enter]  Play Again   [S] Scores   [Q] Quit")

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(0, 1).Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ─────────────────────────────────────────────────────────────────────────────
// Leaderboard
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewLeaderboard(t *ui.Theme) string {
	title := ui.SB(t.HiText).Render("🏆  SCORE HISTORY")
	filter := "All Modes"
	if m.lbFilter != "" {
		filter = strings.ToUpper(m.lbFilter[:1]) + m.lbFilter[1:]
	}
	sub := ui.S(t.DimText).Render("Filter: " + filter)

	hdr := ui.SB(t.Paddle).Render("  #   SCORE   MODE       STREAK   TIME     DATE")
	sep := strings.Repeat("─", 52)

	rows := []string{hdr, sep}
	display := m.scores
	if len(display) > 12 {
		display = display[:12]
	}
	gold := []string{"#FFD700", "#AAAACC", "#CD7F32"}
	for i, r := range display {
		col := t.DimText
		if i < 3 {
			col = gold[i]
		}
		line := fmt.Sprintf("  %-3d %-7d %-10s ×%-7d %-8s %s",
			i+1, r.Score, r.Mode, r.HighStreak,
			fmtSecs(r.DurationSec), r.Timestamp.Format("Jan 02"))
		rows = append(rows, ui.S(col).Render(line))
	}
	if len(m.scores) == 0 {
		rows = append(rows, ui.S(t.DimText).Render("  No scores yet — play a game!"))
	}
	rows = append(rows, sep)

	stats := m.st.Aggregate(m.scores)
	rows = append(rows, ui.S(t.DimText).Render(
		fmt.Sprintf("  Caught: %d · Played: %s · Best ×%d",
			stats.TotalCaught, fmtSecs(stats.TotalTimeSec), stats.BestStreak)))
	rows = append(rows, "")
	rows = append(rows, ui.S(t.DimText).Render(
		"  [0] All  [1] Classic  [2] Arcade  [3] Zen  [4] Timed   [Q] Back"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(1, 2).Render(strings.Join(rows, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, sub, "", box))
}

// ─────────────────────────────────────────────────────────────────────────────
// Help screen
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewHelp(t *ui.Theme) string {
	title := ui.SB(t.Paddle).Render("🏓  PADDLEBALL — HELP")
	c := func(k, v string) string {
		return fmt.Sprintf("  %-22s%s\n",
			ui.SB(t.ScoreText).Render(k),
			ui.S(t.DimText).Render(v))
	}
	content :=
		c("← → / A D", "Move paddle left / right") +
			c("P / Space", "Pause / Resume") +
			c("T", "Cycle colour theme") +
			c("? / H", "Toggle this help") +
			c("R", "Restart (pause / game over)") +
			c("Q / Ctrl+C", "Quit") +
			c("1–4 (title)", "Select game mode") +
			"\n" +
			ui.SB("#FF8C00").Render("  POWER-UPS  (Arcade / Zen)") + "\n" +
			c("Ⓦ Wide Paddle", "Paddle +3 cells for 12s") +
			c("ⓢ Slow Mo", "Ball -35% for 8s") +
			c("ⓕ Fire Paddle", "Score ×2 for 15s") +
			c("ⓘ Iron Shield", "One auto-save (one-time)") +
			c("ⓖ Ghost Ball", "Pass-through once") +
			c("Ⓑ BOMB (dodge!)", "Paddle −2 cells for 10s") +
			"\n" +
			ui.S(t.DimText).Render("  Press any key to go back")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(1, 3).Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, "", box))
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func fmtDur(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func fmtSecs(sec int) string { return fmtDur(time.Duration(sec) * time.Second) }

// visLen approximates the visible (ANSI-stripped) display width of a string.
func visLen(s string) int {
	inEsc := false
	count := 0
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		count += utf8.RuneLen(r)
	}
	return count
}

// ensure termenv and colProfile are referenced so the import is used.
var _ = colProfile