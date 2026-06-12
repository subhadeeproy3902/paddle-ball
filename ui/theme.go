package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ThemeCount is the total number of built-in themes.
const ThemeCount = 4

// Theme holds the full colour palette for one visual style.
type Theme struct {
	Name       string
	Paddle     string
	Ball       string
	Trail      [4]string // index 0=farthest, 3=nearest to ball
	WallTB     string    // top/bottom wall colour
	WallLR     string    // left/right wall colour
	ScoreText  string
	HiText     string
	StreakText  string
	BannerText string
	Border     string
	HeaderBg   string
	FooterBg   string
	PUColor    string
	LivesColor string
	DimText    string
}

// Themes contains all four built-in themes.
var Themes = [ThemeCount]Theme{
	{ // 0 · Neon Arcade (default)
		Name: "Neon",
		Paddle:     "#00FFFF",
		Ball:       "#FFD700",
		Trail:      [4]string{"#4A0000", "#8B1A00", "#FF4500", "#FF8C00"},
		WallTB:     "#4ECDC4",
		WallLR:     "#4ECDC4",
		ScoreText:  "#C3E88D",
		HiText:     "#FFCB6B",
		StreakText:  "#FF5370",
		BannerText: "#FF00FF",
		Border:     "#2D2D44",
		HeaderBg:   "#0A0A1E",
		FooterBg:   "#0A0A1E",
		PUColor:    "#89DDFF",
		LivesColor: "#FF5370",
		DimText:    "#444466",
	},
	{ // 1 · Monochrome
		Name: "Mono",
		Paddle:     "#FFFFFF",
		Ball:       "#FFFFFF",
		Trail:      [4]string{"#1A1A1A", "#333333", "#666666", "#999999"},
		WallTB:     "#AAAAAA",
		WallLR:     "#AAAAAA",
		ScoreText:  "#CCCCCC",
		HiText:     "#FFFFFF",
		StreakText:  "#FFFFFF",
		BannerText: "#FFFFFF",
		Border:     "#333333",
		HeaderBg:   "#0A0A0A",
		FooterBg:   "#0A0A0A",
		PUColor:    "#CCCCCC",
		LivesColor: "#FFFFFF",
		DimText:    "#444444",
	},
	{ // 2 · Sunset
		Name: "Sunset",
		Paddle:     "#FFA07A",
		Ball:       "#FFD700",
		Trail:      [4]string{"#2A0A00", "#5A1500", "#A03020", "#CD5C5C"},
		WallTB:     "#FF6347",
		WallLR:     "#FF6347",
		ScoreText:  "#FFD700",
		HiText:     "#FFF8DC",
		StreakText:  "#FF6B6B",
		BannerText: "#FF8C00",
		Border:     "#4A2000",
		HeaderBg:   "#200800",
		FooterBg:   "#200800",
		PUColor:    "#FFD700",
		LivesColor: "#FF6347",
		DimText:    "#553300",
	},
	{ // 3 · Ocean Night
		Name: "Ocean",
		Paddle:     "#90E0EF",
		Ball:       "#CAF0F8",
		Trail:      [4]string{"#010B1A", "#023E8A", "#0077B6", "#00B4D8"},
		WallTB:     "#0096C7",
		WallLR:     "#0096C7",
		ScoreText:  "#90E0EF",
		HiText:     "#CAF0F8",
		StreakText:  "#48CAE4",
		BannerText: "#ADE8F4",
		Border:     "#023E8A",
		HeaderBg:   "#000814",
		FooterBg:   "#000814",
		PUColor:    "#ADE8F4",
		LivesColor: "#48CAE4",
		DimText:    "#1A3A5A",
	},
}

// ThemeIndexByName returns the index of the named theme (case-insensitive), defaulting to 0.
func ThemeIndexByName(name string) int {
	for i, t := range Themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

// ColorProfile returns the terminal's colour profile (used for fallback decisions).
func ColorProfile() termenv.Profile {
	return termenv.ColorProfile()
}

// S returns a lipgloss style with the given foreground colour.
func S(hex string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}

// SB returns a bold lipgloss style with the given foreground colour.
func SB(hex string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Bold(true)
}

// SBG returns a lipgloss style with foreground and background colours.
func SBG(fg, bg string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Background(lipgloss.Color(bg))
}