package game

// view.go — all rendering.
//
// Aesthetic: restrained dark surfaces, soft off-white text, one accent per
// theme used sparingly (the paddle, the live score, the selected item). No
// background bands — structure comes from a single framed play field and quiet
// hairline rules, so the game reads clean on any terminal background.

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/subhadeeproy3902/paddle-ball/ui"
)

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
	t := &ui.Themes[m.themeIdx]
	if m.width < 60 || m.height < 20 {
		return m.viewTooSmall(t)
	}
	switch m.appPhase {
	case PhaseTitle:
		return m.viewTitle(t)
	case PhaseCountdown:
		return m.viewCountdown(t)
	case PhasePlaying:
		return m.viewPlaying(t)
	case PhasePaused:
		return m.viewPaused(t)
	case PhaseBallLost:
		return m.viewBallLost(t)
	case PhaseGameOver:
		return m.viewGameOver(t)
	case PhaseLeaderboard:
		return m.viewLeaderboard(t)
	case PhaseHelp:
		return m.viewHelp(t)
	}
	return ""
}

func (m Model) viewTooSmall(t *ui.Theme) string {
	return ui.SB(t.Danger).Render(
		fmt.Sprintf("\n  Terminal too small — need at least 60×20\n  Current: %d×%d\n  Please resize.", m.width, m.height))
}

// ─────────────────────────────────────────────────────────────────────────────
// Title screen
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewTitle(t *ui.Theme) string {
	// Compact one-line gradient wordmark, gently shimmering over time.
	off := float64(m.titleFrame) * 0.012
	logo := gradientArt(titleWordmark, t.Accent, t.Phase[4], off)

	tagline := ui.S(t.Faint).Render("a minimalist, physics-based paddle game for your terminal")

	modes := []struct{ label, desc string }{
		{"Classic", "one life · pure score chase"},
		{"Arcade", "three lives · power-ups"},
		{"Zen", "endless rally · no game over"},
		{"Time Trial", "sixty-second sprint"},
	}

	// Menu: a single left-aligned block (marker · number · name · description),
	// centred on the page so every row shares a clean left edge.
	var menu strings.Builder
	for i, mo := range modes {
		sel := i == m.menuSel
		mark := "  "
		numCol, nameStyle, descCol := t.Faint, ui.S(t.Muted), t.Faint
		if sel {
			mark = ui.SB(t.Accent).Render("▸ ")
			numCol, nameStyle, descCol = t.Accent, ui.SB(t.Accent), t.Muted
		}
		row := mark +
			ui.S(numCol).Render(fmt.Sprintf("%d", i+1)) + "  " +
			nameStyle.Render(fmt.Sprintf("%-11s", mo.label)) +
			ui.S(descCol).Render(mo.desc)
		menu.WriteString(row + "\n")
	}
	menuBlock := lipgloss.NewStyle().Width(46).Align(lipgloss.Left).
		Render(strings.TrimRight(menu.String(), "\n"))

	hint := ui.S(t.Faint).Render(
		"↑↓ select   ⏎ start   1-4 quick   S scores   T theme   M sound   ? help")

	dot := ui.S(t.Faint).Render("   ·   ")
	status := ui.S(t.Muted).Render(fmt.Sprintf("best %d", m.hiScore)) + dot +
		ui.S(t.Muted).Render(t.Name+" theme") + dot +
		ui.S(t.Muted).Render("sound "+onOff(m.soundOn))

	body := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		"",
		tagline,
		"",
		"",
		menuBlock,
		"",
		"",
		hint,
		"",
		status,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

// ─────────────────────────────────────────────────────────────────────────────
// Countdown
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewCountdown(t *ui.Theme) string {
	label := "GO"
	if m.countdown > 0 {
		label = fmt.Sprintf("%d", m.countdown)
	}
	big := ui.SB(t.Accent).Render(label)
	sub := ui.S(t.Faint).Render(spaced(strings.ToUpper(m.mode.String())))
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

	if m.bannerTTL > 0 {
		banner := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.bannerColor)).Bold(true).
			Width(m.width).Align(lipgloss.Center).
			Render(spaced(m.bannerText))
		return header + "\n" + banner + "\n" + area + "\n" + footer
	}
	return header + "\n" + area + "\n" + footer
}

func (m Model) buildHeader(t *ui.Theme) string {
	mode := ui.S(t.Faint).Render(spaced(strings.ToUpper(m.mode.String())))
	score := ui.S(t.Faint).Render("SCORE ") + ui.SB(t.Accent).Render(fmt.Sprintf("%d", m.score)) +
		ui.S(t.Faint).Render("    BEST ") + ui.S(t.Muted).Render(fmt.Sprintf("%d", m.hiScore))
	row1 := " " + spread(mode, score, m.width-2)

	// Row 2: phase always; lives / time / streak only when they matter.
	parts := []string{ui.S(t.Phase[m.curPhase.Num-1]).Render(
		fmt.Sprintf("PHASE %d · %s", m.curPhase.Num, m.curPhase.Name))}
	if m.mode == ModeArcade {
		parts = append(parts, ui.S(t.Danger).Render(strings.TrimSpace(strings.Repeat("♥ ", m.lives))))
	}
	if m.mode == ModeTimeTrial {
		remain := m.timeLimit - m.elapsed
		if remain < 0 {
			remain = 0
		}
		secs := int(remain.Seconds())
		tc := t.Muted
		if secs <= 10 {
			tc = t.Danger
		}
		parts = append(parts, ui.S(tc).Render(fmt.Sprintf("%ds left", secs)))
	}
	if m.streak >= 10 {
		parts = append(parts, ui.S(t.Good).Render(fmt.Sprintf("×%d streak", m.streak)))
	}
	row2 := " " + strings.Join(parts, "    ")

	return row1 + "\n" + row2
}

func (m Model) buildFooter(t *ui.Theme) string {
	var left string
	if m.activePU != nil {
		name := ui.S(t.Accent).Render(string(m.activePU.Kind.Glyph()) + " " + m.activePU.Kind.Name())
		if m.activePU.Total > 0 {
			pct := m.activePU.TTL / m.activePU.Total
			left = name + " " + m.puBar.ViewAs(pct) + ui.S(t.Faint).Render(fmt.Sprintf(" %.0fs", m.activePU.TTL))
		} else {
			left = name + ui.S(t.Good).Render(" ready")
		}
	} else {
		left = ui.S(t.Faint).Render("←→ move    P pause    ? help")
	}
	right := ui.S(t.Faint).Render("♪ " + onOff(m.soundOn))
	return " " + spread(left, right, m.width-2)
}

// ─────────────────────────────────────────────────────────────────────────────
// Play-area grid renderer
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) buildPlayArea(t *ui.Theme) string {
	W, H := m.playW, m.playH
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

	// ── Frame ────────────────────────────────────────────────────────────────
	for x := 0; x < W; x++ {
		set(x, 0, '─', t.Wall)
		set(x, H-1, '─', t.Wall)
	}
	for y := 0; y < H; y++ {
		set(0, y, '│', t.Wall)
		set(W-1, y, '│', t.Wall)
	}
	set(0, 0, '╭', t.Wall)
	set(W-1, 0, '╮', t.Wall)
	set(0, H-1, '╰', t.Wall)
	set(W-1, H-1, '╯', t.Wall)

	// ── Ball trail (color ramp carries the fade) ──────────────────────────────
	trailN := len(m.ball.Trail)
	if trailN > len(t.Trail) {
		trailN = len(t.Trail)
	}
	for i := 0; i < trailN; i++ {
		pt := m.ball.Trail[i]
		set(pt.X, pt.Y, '∙', t.Trail[len(t.Trail)-1-i])
	}

	// ── Falling power-ups ──────────────────────────────────────────────────────
	for _, pu := range m.fallingPUs {
		setBold(int(math.Round(pu.X)), int(math.Round(pu.Y)), pu.Kind.Glyph(), pu.Kind.Color())
	}

	// ── Particles ──────────────────────────────────────────────────────────────
	for _, p := range m.particles {
		set(int(math.Round(p.X)), int(math.Round(p.Y)), p.Glyph, p.Color)
	}

	// ── Paddle ─────────────────────────────────────────────────────────────────
	pRow := m.paddleRowY()
	padColor := t.Paddle
	padGlyph := '▬'
	switch {
	case m.paddleFlash > 0:
		padColor = t.Text
		padGlyph = '█'
	case m.activePU != nil && m.activePU.Kind == PUFirePaddle:
		padColor = t.Accent
	case m.shieldActive:
		padColor = t.Good
	}
	px := int(math.Round(m.paddleX))
	for i := 0; i < m.paddleW; i++ {
		setBold(px+i, pRow, padGlyph, padColor)
	}

	// ── Ball ───────────────────────────────────────────────────────────────────
	if m.appPhase == PhasePlaying {
		setBold(int(math.Round(m.ball.X)), int(math.Round(m.ball.Y)), '●', t.Ball)
	}

	// ── Floating score labels ──────────────────────────────────────────────────
	for _, ft := range m.floatTxts {
		fy := int(math.Round(ft.Y))
		fx := int(math.Round(ft.X))
		for i, ch := range ft.Text {
			set(fx+i, fy, ch, ft.Color)
		}
	}

	return renderGrid(grid)
}

// renderGrid flattens the cell grid into a styled string, grouping contiguous
// same-style runs to minimise escape sequences.
func renderGrid(grid [][]cell) string {
	var sb strings.Builder
	for _, row := range grid {
		i := 0
		for i < len(row) {
			c := row[i]
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
	return strings.TrimRight(sb.String(), "\n")
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
	content := ui.SB(t.Accent).Render(spaced("PAUSED")) + "\n\n" +
		kv(t, "Score", fmt.Sprintf("%d", m.score)) +
		kv(t, "Streak", fmt.Sprintf("×%d", m.streak)) +
		kv(t, "Elapsed", fmtDur(m.elapsed.Round(time.Second))) +
		"\n" +
		ui.S(t.Faint).Render("Space / P  resume      R  restart      Q  quit")
	return centerBox(m, t, content)
}

// ─────────────────────────────────────────────────────────────────────────────
// Ball lost — modal (continue? / resuming countdown)
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewBallLost(t *ui.Theme) string {
	ctr := lipgloss.NewStyle().Width(34).Align(lipgloss.Center)
	var lines []string

	if m.lostChoice {
		lines = []string{
			ui.SB(t.Accent).Render(spaced("BALL  OUT")),
			"",
			ui.S(t.Muted).Render(fmt.Sprintf("score %d", m.score)) +
				ui.S(t.Faint).Render(fmt.Sprintf("   ·   best ×%d", m.maxStreak)),
			"",
			ui.S(t.Text).Render("Keep the rally going?"),
			"",
			ui.S(t.Faint).Render("⏎ continue        Q give up"),
		}
	} else {
		head := "BALL OUT"
		if m.mode == ModeArcade {
			head = "YOU LOST A BALL"
		}
		lines = []string{ui.SB(t.Danger).Render(spaced(head)), ""}
		if m.mode == ModeArcade {
			word := "lives"
			if m.lives == 1 {
				word = "life"
			}
			lines = append(lines, ui.S(t.Muted).Render(fmt.Sprintf("%d %s left", m.lives, word)), "")
		}
		n := m.resumeCount
		if n < 1 {
			n = 1
		}
		lines = append(lines,
			ui.S(t.Faint).Render("resuming in"),
			ui.SB(t.Accent).Render(fmt.Sprintf("%d", n)),
			"",
			ui.S(t.Faint).Render("␣ skip        Q give up"),
		)
	}

	content := ctr.Render(strings.Join(lines, "\n"))
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(1, 3).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ─────────────────────────────────────────────────────────────────────────────
// Game Over
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewGameOver(t *ui.Theme) string {
	rank, tier := RankForScore(m.score)
	rankColor := t.Phase[clampInt(tier-1, 0, len(t.Phase)-1)]
	if tier >= len(t.Phase) {
		rankColor = t.Accent
	}
	newBest := m.score > 0 && m.score >= m.hiScore

	width := 40
	rule := ui.S(t.Border).Render(strings.Repeat("─", width))
	content := ui.SB(t.Text).Width(width).Align(lipgloss.Center).Render(spaced("GAME OVER")) + "\n" +
		rule + "\n" +
		kv(t, "Final Score", ui.SB(t.Accent).Render(fmt.Sprintf("%d", m.score))) +
		kv(t, "Rank", ui.SB(rankColor).Render(rank)) +
		kv(t, "Best Streak", ui.S(t.Good).Render(fmt.Sprintf("×%d", m.maxStreak))) +
		kv(t, "Caught / Missed", fmt.Sprintf("%d / %d", m.catches, m.misses)) +
		kv(t, "Top Phase", fmt.Sprintf("%d · %s", m.curPhase.Num, m.curPhase.Name)) +
		kv(t, "Played", fmtDur(m.elapsed.Round(time.Second))) +
		rule
	if newBest {
		content += "\n" + ui.SB(t.Accent).Width(width).Align(lipgloss.Center).Render(spaced("★ NEW PERSONAL BEST"))
	}
	content += "\n\n" + ui.S(t.Faint).Render("R / ⏎  play again      S  scores      Q  quit")
	return centerBox(m, t, content)
}

// ─────────────────────────────────────────────────────────────────────────────
// Leaderboard
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewLeaderboard(t *ui.Theme) string {
	filter := "all modes"
	if m.lbFilter != "" {
		filter = m.lbFilter
	}
	title := ui.SB(t.Text).Render(spaced("SCORE HISTORY")) +
		ui.S(t.Faint).Render("    "+filter)

	width := 56
	rule := ui.S(t.Border).Render(strings.Repeat("─", width))
	rows := []string{
		title, "",
		ui.S(t.Faint).Render(fmt.Sprintf("  %-3s %-7s %-10s %-8s %-8s %s", "#", "SCORE", "MODE", "STREAK", "TIME", "DATE")),
		rule,
	}

	display := m.scores
	if len(display) > 12 {
		display = display[:12]
	}
	for i, r := range display {
		col := t.Muted
		if i == 0 {
			col = t.Accent
		} else if i >= 3 {
			col = t.Faint
		}
		line := fmt.Sprintf("  %-3d %-7d %-10s ×%-7d %-8s %s",
			i+1, r.Score, r.Mode, r.HighStreak, fmtSecs(r.DurationSec), r.Timestamp.Format("Jan 02"))
		rows = append(rows, ui.S(col).Render(line))
	}
	if len(m.scores) == 0 {
		rows = append(rows, ui.S(t.Faint).Render("  no scores yet — play a game"))
	}
	rows = append(rows, rule)

	stats := m.st.Aggregate(m.scores)
	rows = append(rows, ui.S(t.Muted).Render(fmt.Sprintf("  caught %d   ·   played %s   ·   best ×%d",
		stats.TotalCaught, fmtSecs(stats.TotalTimeSec), stats.BestStreak)))
	rows = append(rows, "")

	if m.confirmClear {
		rows = append(rows, ui.SB(t.Danger).Render("  press C again to erase ALL history · any key cancels"))
	} else {
		rows = append(rows, ui.S(t.Faint).Render("  0 all · 1-4 filter · C clear · Q back"))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(1, 2).Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ─────────────────────────────────────────────────────────────────────────────
// Help screen
// ─────────────────────────────────────────────────────────────────────────────

func (m Model) viewHelp(t *ui.Theme) string {
	c := func(k, v string) string {
		return "  " + ui.S(t.Muted).Render(fmt.Sprintf("%-16s", k)) + ui.S(t.Faint).Render(v) + "\n"
	}
	content := ui.SB(t.Text).Render(spaced("HOW TO PLAY")) + "\n\n" +
		c("← → / A D", "move the paddle") +
		c("drag mouse", "hold left button to move") +
		c("P / Space", "pause and resume") +
		c("T", "cycle color theme") +
		c("M", "toggle sound") +
		c("R", "restart (pause / game over)") +
		c("Q / Ctrl+C", "quit") +
		"\n" + ui.S(t.Muted).Render("  POWER-UPS") + ui.S(t.Faint).Render("  (Arcade / Zen)") + "\n" +
		c("Ⓦ Wide", "paddle +3 cells, 12s") +
		c("ⓢ Slow", "ball −40%, 8s") +
		c("ⓕ Fire", "double score, 15s") +
		c("ⓘ Shield", "one automatic save") +
		c("ⓖ Ghost", "phase through once") +
		c("Ⓑ Bomb", "paddle −2 cells, 10s — dodge it") +
		"\n" + ui.S(t.Faint).Render("  press any key to go back")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(1, 3).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared helpers
// ─────────────────────────────────────────────────────────────────────────────

// kv renders one "label   value" row used in the pause / game-over panels.
func kv(t *ui.Theme, label, val string) string {
	return "  " + ui.S(t.Faint).Render(fmt.Sprintf("%-16s", label)) + val + "\n"
}

// centerBox wraps content in a hairline rounded border and centres it.
func centerBox(m Model, t *ui.Theme, content string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		Padding(1, 4).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// spread places left and right text at the two ends of a row of the given width.
func spread(left, right string, width int) string {
	gap := width - visLen(left) - visLen(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// hexToRGB parses "#rrggbb" into 0–255 components.
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// gradientArt colorises ASCII art with an animated horizontal gradient between
// two theme colors. It emits truecolor ANSI per glyph (spaces left bare) — cheap
// to rebuild each frame and only used on the title screen. `offset` shifts the
// gradient phase to make it shimmer over time.
func gradientArt(lines []string, hexA, hexB string, offset float64) string {
	r1, g1, b1 := hexToRGB(hexA)
	r2, g2, b2 := hexToRGB(hexB)
	maxW := 1
	for _, l := range lines {
		if len(l) > maxW {
			maxW = len(l)
		}
	}
	lerp := func(a, b int, t float64) int { return a + int((float64(b-a))*t) }
	var sb strings.Builder
	for li, line := range lines {
		for x := 0; x < len(line); x++ {
			ch := line[x]
			if ch == ' ' {
				sb.WriteByte(' ')
				continue
			}
			t := float64(x)/float64(maxW-1) + offset
			t -= math.Floor(t)        // wrap to [0,1)
			if tw := t * 2; tw <= 1 { // triangle wave → smooth back-and-forth
				t = tw
			} else {
				t = 2 - tw
			}
			fmt.Fprintf(&sb, "\x1b[1;38;2;%d;%d;%dm%c", lerp(r1, r2, t), lerp(g1, g2, t), lerp(b1, b2, t), ch)
		}
		sb.WriteString("\x1b[0m")
		if li < len(lines)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// spaced inserts thin spacing between letters for a quiet, editorial caps look.
func spaced(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func fmtDur(d time.Duration) string {
	mm := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if mm > 0 {
		return fmt.Sprintf("%dm%02ds", mm, s)
	}
	return fmt.Sprintf("%ds", s)
}

func fmtSecs(sec int) string { return fmtDur(time.Duration(sec) * time.Second) }

// visLen approximates the visible (ANSI-stripped) column width of a string.
// Each non-escape rune counts as one cell — correct for the BMP glyphs used in
// the HUD (box-drawing, ♥, ·, ▸); we avoid wide emoji in aligned rows.
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
		count++
	}
	return count
}
