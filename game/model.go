package game

// model.go — all game types, Model struct, constructor, and Init.
// Uses:
//   bubbletea  – TUI framework
//   harmonica  – spring physics for paddle
//   bubbles    – progress bar for active power-up timer
//   lipgloss   – imported indirectly via ui package
//   cobra      – CLI entry (in cmd/)
//   termenv    – colour profile (in ui/)

import (
	"math"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/harmonica"
	"github.com/subhadeeproy3902/paddle-ball/store"
	"github.com/subhadeeproy3902/paddle-ball/ui"
)

// ─────────────────────────────────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────────────────────────────────

const (
	BaseSpeed       = 22.0 // cells/second at Phase 1
	KeyMoveStep     = 4.0  // paddle target cells per key event
	PaddleRow       = 3    // rows from bottom of play area
	SpringFreq      = 22.0 // harmonica angular frequency
	SpringDamp      = 0.65 // harmonica damping ratio (< 1 = satisfying overshoot)
	PUCatchInterval = 7    // catches between power-up spawns (Arcade/Zen)
	KeyAutoRelease  = 150  // ms of no key → stop paddle movement
)

// ─────────────────────────────────────────────────────────────────────────────
// Enums
// ─────────────────────────────────────────────────────────────────────────────

type TickMsg time.Time

type AppPhase int

const (
	PhaseTitle       AppPhase = iota
	PhaseCountdown
	PhasePlaying
	PhasePaused
	PhaseGameOver
	PhaseLeaderboard
	PhaseHelp
)

type GameMode int

const (
	ModeClassic  GameMode = iota
	ModeArcade
	ModeZen
	ModeTimeTrial
)

func (m GameMode) String() string {
	return [...]string{"Classic", "Arcade", "Zen", "Time Trial"}[m]
}
func (m GameMode) Code() string {
	return [...]string{"classic", "arcade", "zen", "timed"}[m]
}

// ─────────────────────────────────────────────────────────────────────────────
// Game object types
// ─────────────────────────────────────────────────────────────────────────────

// Pt is a 2-D integer grid coordinate.
type Pt struct{ X, Y int }

// Ball — position, velocity, and trail history.
// Game is VERTICAL: paddle at bottom (horizontal), ball travels up/down.
type Ball struct {
	X, Y   float64 // sub-cell position (rendered at nearest int)
	VX, VY float64 // velocity in cells/second (VY positive = downward)
	Trail  []Pt    // last N integer positions, index 0 = most recent
}

// Particle — one transient visual element.
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64 // 1.0 → 0.0
	Decay  float64 // life units lost per second
	Glyph  rune
	Color  string
}

// FloatText — a "+N" score label that drifts upward and fades.
type FloatText struct {
	X, Y  float64
	Text  string
	Color string
	Life  float64
	Decay float64
}

// PowerUpKind identifies a power-up type.
type PowerUpKind int

const (
	PUWidePaddle PowerUpKind = iota
	PUSlowMo
	PUFirePaddle
	PUIronShield
	PUGhost
	PUBomb
)

func (k PowerUpKind) Glyph() rune {
	return []rune{'Ⓦ', 'ⓢ', 'ⓕ', 'ⓘ', 'ⓖ', 'Ⓑ'}[k]
}
func (k PowerUpKind) Name() string {
	return []string{"Wide Paddle", "Slow Mo", "Fire Paddle", "Iron Shield", "Ghost Ball", "BOMB"}[k]
}
func (k PowerUpKind) Duration() float64 { return []float64{12, 8, 15, 0, 0, 10}[k] }
func (k PowerUpKind) Color() string {
	return []string{"#89DDFF", "#C3E88D", "#FFD700", "#4ECDC4", "#FFFFFF", "#FF5370"}[k]
}

// FallingPU is a power-up falling from the top of the play area.
type FallingPU struct {
	X, Y     float64
	Kind     PowerUpKind
	VY       float64
	AnimStep float64
	AnimTTL  float64
}

// ActivePU is a power-up currently in effect.
type ActivePU struct {
	Kind  PowerUpKind
	TTL   float64 // remaining seconds; 0 = one-shot consumed
	Total float64 // original duration (for progress bar)
}

// Phase — difficulty level configuration.
type Phase struct {
	Num       int
	Name      string
	Color     string
	Emoji     string
	MinScore  int
	SpeedMult float64 // multiplied by BaseSpeed
	PaddleW   int     // paddle width in cells
	TrailLen  int     // trail positions to show
}

// Phases lists all five difficulty levels in ascending order.
var Phases = []Phase{
	{1, "Warm Up",    "#C3E88D", "🌱", 0,   1.00, 14, 1},
	{2, "Heating Up", "#FFCB6B", "🔥", 10,  1.25, 12, 2},
	{3, "On Fire",    "#FF8C00", "💥", 25,  1.55, 10, 3},
	{4, "Blazing",    "#FF5370", "⚡", 50,  1.90, 8,  4},
	{5, "INSANE",     "#FF00FF", "🏆", 100, 2.35, 6,  5},
}

// PhaseForScore returns the Phase that best matches the given score.
func PhaseForScore(score int) Phase {
	p := Phases[0]
	for _, ph := range Phases {
		if score >= ph.MinScore {
			p = ph
		}
	}
	return p
}

// ─────────────────────────────────────────────────────────────────────────────
// Model
// ─────────────────────────────────────────────────────────────────────────────

// Model is the top-level bubbletea model; it holds ALL application state.
type Model struct {
	// ── window dimensions ──────────────────────────────────────────────────
	width, height int
	playW, playH  int // play-area dimensions (inside header/footer/walls)

	// ── screen state machine ───────────────────────────────────────────────
	appPhase AppPhase
	mode     GameMode
	menuSel  int // title-screen cursor (0–3)

	// ── ball ───────────────────────────────────────────────────────────────
	ball Ball

	// ── paddle (spring-driven) ─────────────────────────────────────────────
	//   Paddle is HORIZONTAL at the bottom of the play area.
	//   paddleX = left edge of paddle (float for spring).
	//   Spring drives paddleX toward paddleTargX.
	paddleX      float64
	paddleVX     float64          // spring velocity (used internally by harmonica)
	paddleTargX  float64          // target position (set by key events)
	paddleW      int              // width in cells (changes with phase)
	paddleLastVX float64          // velocity for spin transfer
	paddleFlash  float64          // seconds of white-flash remaining after a hit
	paddleSpring harmonica.Spring // spring model (harmonica)

	// ── key direction (moved directly per key event, no held-state) ────────
	// We shift paddleTargX in handleKey; the spring does the rest.
	// paddleDir tracks the last direction for optional auto-decel logic.
	paddleDir   float64   // -1, 0, +1
	lastKeyTime time.Time // when the last paddle key was received

	// ── game objects ───────────────────────────────────────────────────────
	particles  []Particle
	floatTxts  []FloatText
	fallingPUs []FallingPU
	activePU   *ActivePU

	// ── special power-up states ────────────────────────────────────────────
	shieldActive bool
	ghostActive  bool

	// ── scoring ────────────────────────────────────────────────────────────
	score     int
	hiScore   int
	streak    int
	maxStreak int
	catches   int
	misses    int
	catchesSinceLastPU int

	// ── difficulty ─────────────────────────────────────────────────────────
	curPhase Phase

	// ── lives (Arcade mode) ────────────────────────────────────────────────
	lives int

	// ── timing ─────────────────────────────────────────────────────────────
	lastTick  time.Time
	gameStart time.Time
	elapsed   time.Duration
	timeLimit time.Duration // Time Trial only

	// ── countdown (3 … 2 … 1 … GO!) ───────────────────────────────────────
	countdown int
	cdTTL     float64

	// ── phase-transition banner ────────────────────────────────────────────
	bannerText  string
	bannerColor string
	bannerTTL   float64

	// ── bubbles progress bar for active power-up ───────────────────────────
	puBar progress.Model

	// ── theme ──────────────────────────────────────────────────────────────
	themeIdx int

	// ── store / leaderboard ────────────────────────────────────────────────
	st       *store.Store
	scores   []store.ScoreRecord
	lbFilter string
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// NewModel creates a fresh model pre-loading the hi-score and saved config.
func NewModel(modeStr, themeStr string) Model {
	st := store.New()
	cfg := st.LoadConfig()

	themeIdx := cfg.ThemeIndex
	if themeStr != "" {
		themeIdx = ui.ThemeIndexByName(themeStr)
	}

	mode := ModeClassic
	jumpToGame := false
	if modeStr != "" {
		mode = parseModeStr(modeStr)
		jumpToGame = true
	}

	bar := progress.New(
		progress.WithGradient("#C3E88D", "#FF5370"),
		progress.WithWidth(14),
		progress.WithoutPercentage(),
	)

	spring := harmonica.NewSpring(harmonica.FPS(60), SpringFreq, SpringDamp)

	m := Model{
		appPhase:     PhaseTitle,
		mode:         mode,
		menuSel:      0,
		themeIdx:     themeIdx,
		puBar:        bar,
		paddleSpring: spring,
		st:           st,
		hiScore:      st.HiScore(""),
	}

	if jumpToGame {
		m.startCountdown()
	}
	return m
}

func parseModeStr(s string) GameMode {
	switch s {
	case "arcade":
		return ModeArcade
	case "zen":
		return ModeZen
	case "timed", "time-trial":
		return ModeTimeTrial
	}
	return ModeClassic
}

// ─────────────────────────────────────────────────────────────────────────────
// bubbletea interface
// ─────────────────────────────────────────────────────────────────────────────

// Init is called once when the program starts; it fires the first tick and
// initialises the bubbles progress bar.
func (m Model) Init() tea.Cmd {
	return tea.Batch(schedTick(), m.puBar.Init())
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal game-state transitions
// ─────────────────────────────────────────────────────────────────────────────

func (m *Model) startCountdown() {
	m.appPhase = PhaseCountdown
	m.countdown = 3
	m.cdTTL = 1.0
	m.resetAll()
}

func (m *Model) startPlaying() {
	m.appPhase = PhasePlaying
	m.gameStart = time.Now()
	m.lastTick = time.Now()
	m.elapsed = 0
	m.score = 0
	m.streak = 0
	m.maxStreak = 0
	m.catches = 0
	m.misses = 0
	m.particles = nil
	m.floatTxts = nil
	m.fallingPUs = nil
	m.activePU = nil
	m.shieldActive = false
	m.ghostActive = false
	m.catchesSinceLastPU = 0
	m.bannerTTL = 0
	m.curPhase = Phases[0]

	switch m.mode {
	case ModeArcade:
		m.lives = 3
	case ModeTimeTrial:
		m.lives = 999
		m.timeLimit = 60 * time.Second
	default:
		m.lives = 1
	}

	m.hiScore = m.st.HiScore(m.mode.Code())
	m.resetAll()
}

// resetAll re-positions ball and paddle to neutral starting positions.
func (m *Model) resetAll() {
	if m.playH == 0 {
		m.playH = 18
		m.playW = 72
	}
	ph := PhaseForScore(m.score)
	m.curPhase = ph
	m.paddleW = ph.PaddleW

	// Paddle centred horizontally at bottom
	cx := float64(m.playW)/2 - float64(m.paddleW)/2
	m.paddleX = cx
	m.paddleVX = 0
	m.paddleTargX = cx
	m.paddleFlash = 0

	// Ball starts at roughly 1/3 from top, angled downward
	bx := float64(m.playW) / 2
	by := float64(m.playH) / 3
	speed := BaseSpeed * ph.SpeedMult
	// Random downward angle: vy positive (downward), vx small random
	angle := (rand.Float64() - 0.5) * math.Pi / 2.5 // ±36°
	m.ball = Ball{
		X:  bx,
		Y:  by,
		VX: speed * math.Sin(angle),
		VY: speed * math.Cos(angle), // positive = moving toward paddle
	}
}

// schedTick schedules the next game tick (~60fps).
func schedTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}