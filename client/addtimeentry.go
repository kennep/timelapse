package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getTimeEntriesCmd)
}

var addTimeEntryCmd = &cobra.Command{
	Use:   "add-entry PROJECTNAME START END",
	Short: "Add time entry",
	Long:  `Add a time entry for a given project`,
	Args:  cobra.RangeArgs(2, 3),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return addTimeEntry(args[0])
		} else {
			return addTimeEntry("")
		}
	}),
}

func addTimeEntry(projectName string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	timeentries, err := apiClient.GetTimeEntries(projectName)
	if err != nil {
		return err
	}
	for _, timeentry := range timeentries {
		fmt.Println(timeentry)
	}

	return nil
}
