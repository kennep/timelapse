package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getTimeEntriesCmd)
}

var getTimeEntriesCmd = &cobra.Command{
	Use:   "get-entries [PROJECTNAME]",
	Short: "Show time entries",
	Long:  `List time entries for a project, or for all projects`,
	Args:  cobra.MaximumNArgs(1),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return getTimeEntries(args[0])
		} else {
			return getTimeEntries("")
		}
	}),
}

func getTimeEntries(projectName string) error {
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
