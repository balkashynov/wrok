package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
)

var archiveCmd = &cobra.Command{
	Use:     "archive [task-id]",
	Aliases: []string{"a"},
	Short:   "Archive a task",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Printf("Error: invalid task ID '%s'\n", args[0])
			return
		}

		task, err := db.ArchiveTask(uint(taskID))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("🗃️  Archived task #%d: %s\n", task.ID, task.Title)
		if task.ArchivedAt != nil {
			fmt.Printf("Archived at: %s\n", task.ArchivedAt.Format("15:04:05"))
		}
	},
}

var unarchiveCmd = &cobra.Command{
	Use:     "unarchive [task-id]",
	Aliases: []string{"ua"},
	Short:   "Unarchive a task (move back to todo)",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Printf("Error: invalid task ID '%s'\n", args[0])
			return
		}

		task, err := db.UnarchiveTask(uint(taskID))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("📤 Unarchived task #%d: %s\n", task.ID, task.Title)
		fmt.Printf("Status: %s\n", task.Status)
	},
}