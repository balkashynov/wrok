package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
)

var doneCmd = &cobra.Command{
	Use:   "done [task-id]",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Printf("Error: invalid task ID '%s'\n", args[0])
			return
		}

		task, err := db.MarkTaskDone(uint(taskID))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✅ Marked task #%d as done: %s\n", task.ID, task.Title)
		if task.DoneAt != nil {
			fmt.Printf("Completed at: %s\n", task.DoneAt.Format("15:04:05"))
		}
	},
}