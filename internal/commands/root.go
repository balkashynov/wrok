package commands

import (
	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "wrok",
	Short: "A CLI todo and time tracker",
	Long: `wrok is a command-line tool that combines task management with time tracking.
Track your tasks, monitor your time, and generate reports all from the terminal.`,
}

// initDB initializes the database and panics on error
func initDB() {
	if err := db.Initialize(); err != nil {
		panic(err) // For now, panic on DB init failure
	}
}

// withDB wraps a command function to initialize the database first
func withDB(fn func(*cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		initDB()
		fn(cmd, args)
	}
}

// SetVersion sets the version information
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(undoneCmd)
	rootCmd.AddCommand(archiveCmd)
	rootCmd.AddCommand(unarchiveCmd)
	rootCmd.AddCommand(jiraCmd)
	rootCmd.AddCommand(helpCmd)
	rootCmd.AddCommand(versionCmd)
}