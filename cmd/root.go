package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/subhadeeproy3902/paddle-ball/game"
	"github.com/subhadeeproy3902/paddle-ball/store"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

// Execute wires all sub-commands and runs cobra.
func Execute(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date

	root := &cobra.Command{
		Use:   "paddle-ball",
		Short: "A minimalist physics-based terminal paddleball game",
		Long: lipgloss.NewStyle().Foreground(lipgloss.Color("#cc785c")).Bold(true).
			Render("paddle-ball") +
			" — a minimalist paddleball game for the terminal.\n" +
			"Sub-stepped physics · spring paddle · five themes · score history",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGame("", "")
		},
	}

	// ── play ─────────────────────────────────────────────────────────────
	var playMode, playTheme string
	playCmd := &cobra.Command{
		Use:     "play",
		Short:   "Start a game immediately",
		Example: "  paddle-ball play\n  paddle-ball play --mode arcade\n  paddle-ball play --mode timed --theme nord",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGame(playMode, playTheme)
		},
	}
	playCmd.Flags().StringVarP(&playMode, "mode", "m", "", "Game mode: classic | arcade | zen | timed")
	playCmd.Flags().StringVarP(&playTheme, "theme", "t", "", "Theme: claude | mono | nord | moss | ember")

	// ── scores ───────────────────────────────────────────────────────────
	var scoresMode string
	var scoresAll, scoresJSON bool
	scoresCmd := &cobra.Command{
		Use:   "scores",
		Short: "View score history and leaderboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			st := store.New()
			records, err := st.LoadAll()
			if err != nil {
				return fmt.Errorf("could not load scores: %w", err)
			}
			if len(records) == 0 {
				fmt.Println("No scores yet. Run `paddle-ball play` first!")
				return nil
			}
			if scoresMode != "" {
				filtered := records[:0]
				for _, r := range records {
					if r.Mode == scoresMode {
						filtered = append(filtered, r)
					}
				}
				records = filtered
			}
			top := records
			if !scoresAll && len(top) > 10 {
				top = top[:10]
			}
			if scoresJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(top)
			}
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#cc785c")).Bold(true)
			fmt.Println(style.Render("paddle-ball — score history"))
			fmt.Println()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "#\tSCORE\tMODE\tSTREAK\tDURATION\tDATE")
			fmt.Fprintln(w, "─\t─────\t────\t──────\t────────\t────")
			for i, r := range top {
				dur := time.Duration(r.DurationSec) * time.Second
				fmt.Fprintf(w, "%d\t%d\t%s\t×%d\t%s\t%s\n",
					i+1, r.Score, r.Mode, r.HighStreak,
					fmtD(dur), r.Timestamp.Format("Jan 02 '06"))
			}
			w.Flush()
			stats := st.Aggregate(records)
			fmt.Printf("\nAll-time: caught %d · played %s · best ×%d\n",
				stats.TotalCaught, fmtD(time.Duration(stats.TotalTimeSec)*time.Second), stats.BestStreak)
			return nil
		},
	}
	scoresCmd.Flags().StringVarP(&scoresMode, "mode", "m", "", "Filter by mode: classic | arcade | zen | timed")
	scoresCmd.Flags().BoolVarP(&scoresAll, "all", "a", false, "Show full history (not just top 10)")
	scoresCmd.Flags().BoolVar(&scoresJSON, "json", false, "Print raw JSON to stdout")

	// ── reset ────────────────────────────────────────────────────────────
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Delete all saved scores (prompts for confirmation)",
		RunE: func(cmd *cobra.Command, args []string) error {
			st := store.New()
			records, _ := st.LoadAll()
			fmt.Printf("This will permanently delete %d record(s).\n", len(records))
			fmt.Print("Type YES to confirm: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "YES" {
				fmt.Println("Aborted.")
				return nil
			}
			if err := st.Reset(); err != nil {
				return fmt.Errorf("reset failed: %w", err)
			}
			fmt.Println("All scores deleted.")
			return nil
		},
	}

	// ── version ──────────────────────────────────────────────────────────
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("paddle-ball %s\n  commit: %s\n  built:  %s\n",
				buildVersion, buildCommit, buildDate)
		},
	}

	root.AddCommand(playCmd, scoresCmd, resetCmd, versionCmd)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runGame(mode, theme string) error {
	m := game.NewModel(mode, theme)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("game error: %w", err)
	}
	return nil
}

func fmtD(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
