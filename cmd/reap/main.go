package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/legostin/reap/internal/config"
	"github.com/legostin/reap/internal/ports"
	"github.com/legostin/reap/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "reap",
	Short: "Interactive TUI for viewing and killing processes on ports",
	Long:  "reap — like htop meets lsof. View listening ports, filter, sort, and kill processes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		scanner := ports.NewScanner()
		model := tui.New(scanner, cfg)
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func main() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(killCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
