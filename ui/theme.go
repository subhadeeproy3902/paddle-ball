package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Theme holds one restrained, dark color palette.
//
// Design philosophy (see DESIGN-claude.md): a single warm/cool accent does the
// talking, text is a soft off-white, and everything else is a quiet ramp of
// muted tones. Color is scarce on purpose — the accent only appears on the ball,
// the live score, and momentary highlights. No rainbow.
type Theme struct {
	Name string

	// ── surfaces ──────────────────────────────────────────────────────────
	Bg     string // chrome background (header / footer bands)
	Border string // hairline border tone

	// ── text ──────────────────────────────────────────────────────────────
	Text  string // primary on-dark text
	Muted string // secondary labels
	Faint string // hints, fine print, inactive items

	// ── accents (used sparingly) ──────────────────────────────────────────
	Accent string // the one signature color
	Good   string // restrained positive (streak / success)
	Danger string // restrained negative (miss / lives / warnings)

	// ── game objects ──────────────────────────────────────────────────────
	Ball   string
	Paddle string
	Wall   string
	Trail  [4]string // 0 = farthest/faintest … 3 = nearest to ball

	// ── difficulty ramp ───────────────────────────────────────────────────
	// Five quiet shades, calm → hot, drawn from the theme's own family so
	// phase labels never break the palette.
	Phase [5]string
}

// HeaderBg / FooterBg expose the chrome background under the names the renderer
// expects; both share Bg so the chrome reads as one continuous band.
func (t *Theme) HeaderBg() string { return t.Bg }
func (t *Theme) FooterBg() string { return t.Bg }

// Themes contains all built-in palettes. Index 0 is the default.
var Themes = []Theme{
	{ // 0 · Claude — warm dark, coral accent (default)
		Name:   "Claude",
		Bg:     "#1f1e1b",
		Border: "#33312c",
		Text:   "#ece6da",
		Muted:  "#a09d96",
		Faint:  "#6c6a64",
		Accent: "#cc785c",
		Good:   "#5db8a6",
		Danger: "#c87b6a",
		Ball:   "#faf9f5",
		Paddle: "#cc785c",
		Wall:   "#46443e",
		Trail:  [4]string{"#3a2a22", "#5c4034", "#915541", "#c0795f"},
		Phase:  [5]string{"#8e8b82", "#b08a72", "#cc785c", "#d98a5f", "#e8a55a"},
	},
	{ // 1 · Mono — pure grayscale, zero hue
		Name:   "Mono",
		Bg:     "#121212",
		Border: "#2a2a2a",
		Text:   "#e8e8e8",
		Muted:  "#9a9a9a",
		Faint:  "#5a5a5a",
		Accent: "#f5f5f5",
		Good:   "#cfcfcf",
		Danger: "#8a8a8a",
		Ball:   "#ffffff",
		Paddle: "#e8e8e8",
		Wall:   "#3a3a3a",
		Trail:  [4]string{"#222222", "#3d3d3d", "#5e5e5e", "#888888"},
		Phase:  [5]string{"#6a6a6a", "#8a8a8a", "#a8a8a8", "#cccccc", "#f5f5f5"},
	},
	{ // 2 · Nord — cool slate, frost accent
		Name:   "Nord",
		Bg:     "#20242c",
		Border: "#333b48",
		Text:   "#e5e9f0",
		Muted:  "#9aa4b8",
		Faint:  "#5b647a",
		Accent: "#88c0d0",
		Good:   "#a3be8c",
		Danger: "#bf616a",
		Ball:   "#eceff4",
		Paddle: "#88c0d0",
		Wall:   "#3b4252",
		Trail:  [4]string{"#2e3440", "#3b4252", "#4c566a", "#6a7791"},
		Phase:  [5]string{"#7a89a8", "#81a1c1", "#88c0d0", "#8fbcbb", "#a3be8c"},
	},
	{ // 3 · Moss — warm forest, sage accent
		Name:   "Moss",
		Bg:     "#1b1f1a",
		Border: "#313a2e",
		Text:   "#e6e8df",
		Muted:  "#9aa48f",
		Faint:  "#5f6a55",
		Accent: "#a3b18a",
		Good:   "#8aa872",
		Danger: "#c08a6a",
		Ball:   "#f0efe4",
		Paddle: "#a3b18a",
		Wall:   "#3a4434",
		Trail:  [4]string{"#26301f", "#3a4a30", "#566b45", "#7a9460"},
		Phase:  [5]string{"#8b9a7c", "#9aab85", "#a3b18a", "#bcae74", "#d4a017"},
	},
	{ // 4 · Ember — near-black, single crimson accent
		Name:   "Ember",
		Bg:     "#1a1716",
		Border: "#332a27",
		Text:   "#ece4e1",
		Muted:  "#a0918c",
		Faint:  "#675a56",
		Accent: "#d9694f",
		Good:   "#c9a26b",
		Danger: "#d9694f",
		Ball:   "#f7efe9",
		Paddle: "#d9694f",
		Wall:   "#43332e",
		Trail:  [4]string{"#33201a", "#552f24", "#8a4734", "#c25f44"},
		Phase:  [5]string{"#9a8079", "#b87a63", "#d9694f", "#e08158", "#eaa15c"},
	},
}

// ThemeCount is the number of built-in themes.
var ThemeCount = len(Themes)

// ThemeIndexByName returns the index of the named theme (case-insensitive-ish),
// defaulting to 0 (Claude).
func ThemeIndexByName(name string) int {
	for i, t := range Themes {
		if eqFold(t.Name, name) {
			return i
		}
	}
	return 0
}

func eqFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// ColorProfile returns the terminal's color profile (used for fallback checks).
func ColorProfile() termenv.Profile { return termenv.ColorProfile() }

// S returns a lipgloss style with the given foreground color.
func S(hex string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}

// SB returns a bold lipgloss style with the given foreground color.
func SB(hex string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Bold(true)
}

// SBG returns a lipgloss style with foreground and background colors.
func SBG(fg, bg string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Background(lipgloss.Color(bg))
}
