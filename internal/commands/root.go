package commands

import (
	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
)

var rootCmd = &cobra.Command{
	Use:   "wrok",
	Short: "A CLI todo and time tracker",
	Long: `wrok is a command-line tool that combines task management with time tracking.
Track your tasks, monitor your time, and generate reports all from the terminal.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize database before any command runs
		if err := db.Initialize(); err != nil {
			panic(err) // For now, panic on DB init failure
		}
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(undoneCmd)
	rootCmd.AddCommand(archiveCmd)
	rootCmd.AddCommand(unarchiveCmd)
}