package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"nvim-quest/internal/content"
	"nvim-quest/internal/progress"
	"nvim-quest/internal/ui"
	"nvim-quest/internal/update"
)

// build info, injected from main via Execute.
var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var noUpdateCheck bool

var rootCmd = &cobra.Command{
	Use:   "nvim-quest",
	Short: "Learn Neovim through an epic three-act terminal quest",
	RunE: func(cmd *cobra.Command, args []string) error {
		lessons, err := content.All()
		if err != nil {
			return err
		}
		path := progress.DefaultPath()
		prog := progress.Load(path)
		p := tea.NewProgram(
			ui.New(lessons, prog, path, buildVersion, updateChecker()),
			tea.WithAltScreen(),
		)
		_, err = p.Run()
		return err
	},
}

// Execute runs the CLI. version/commit/date are the build-time values from main.
func Execute(version, commit, date string) {
	buildVersion, buildCommit, buildDate = version, commit, date
	rootCmd.Version = version
	rootCmd.PersistentFlags().BoolVar(&noUpdateCheck, "no-update-check", false,
		"Disable the launch-time check for a newer release")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// updateChecker returns a function the TUI runs (off the render path) to learn
// whether a newer release exists. It returns "" when checking is disabled, the
// 24h throttle hasn't elapsed and nothing is cached, on any error, or when the
// build is already current — i.e. it returns a version string ONLY when there is
// a strictly newer release to advertise. Returns nil when checking is disabled.
func updateChecker() func() string {
	if !update.ShouldCheck(noUpdateCheck, os.Getenv("NVIM_QUEST_NO_UPDATE_CHECK"), buildVersion) {
		return nil
	}
	version := buildVersion
	return func() string {
		path, err := update.DefaultCachePath()
		if err != nil {
			return ""
		}
		c := update.Load(path)
		latest := c.LatestVersion
		if update.Due(c, time.Now(), update.CheckInterval) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if v, err := update.Latest(ctx); err == nil && v != "" {
				latest = v
				_ = update.Save(path, update.Cache{LastCheck: time.Now(), LatestVersion: v})
			}
		}
		if update.Newer(version, latest) {
			return latest
		}
		return ""
	}
}

// runUpdate performs the self-update and prints the outcome. Shared by the
// `update` subcommand.
func runUpdate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	newV, err := update.Apply(ctx, buildVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not update: %v.\nIf you installed via go install or a package manager, update with that instead; otherwise re-run with sufficient permissions (e.g. sudo).\n", err)
		return err
	}
	if newV == "" {
		fmt.Fprintf(os.Stderr, "nvim-quest is up to date (%s).\n", buildVersion)
		return nil
	}
	fmt.Fprintf(os.Stderr, "Updated nvim-quest %s → %s.\n", buildVersion, newV)
	return nil
}
