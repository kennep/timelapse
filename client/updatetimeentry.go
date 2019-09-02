package client

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateTimeEntryCmd)
	updateTimeEntryCmd.Flags().StringVarP(&startTime, "start", "s", "", "Entry start time/date")
	updateTimeEntryCmd.Flags().StringVarP(&endTime, "end", "e", "", "Entry end time/date")
	updateTimeEntryCmd.Flags().StringVarP(&breaks, "breaks", "b", "", "Break duration")
	updateTimeEntryCmd.Flags().StringVarP(&entryType, "type", "t", "work", "Entry type (work|sick|sick-child|vacation)")
	updateTimeEntryCmd.Flags().StringVarP(&entryComment, "comment", "c", "", "Entry comment")
}

var updateTimeEntryCmd = &cobra.Command{
	Use:   "update-entry PROJECTNAME ENTRYID [-s START] [-e END]",
	Short: "Update time entry",
	Long:  `Update a time entry for a given project`,
	Args:  cobra.ExactArgs(2),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return updateTimeEntry(cmd, args[0], args[1])
	}),
}

func updateTimeEntry(cmd *cobra.Command, projectName string, entryID string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	entry, err := apiClient.GetTimeEntry(projectName, entryID)
	if err != nil {
		return err
	}

	if startTime != "" {
		entry.Start, err = ParseTimeRef(startTime, entry.Start)
		if err != nil {
			return err
		}
	}

	if endTime != "" {
		entry.End, err = ParseTimeRef(endTime, entry.End)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("breaks") {
		entry.Breaks, err = time.ParseDuration(breaks)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("type") {
		entry.Type = entryType
	}

	if cmd.Flags().Changed("comment") {
		entry.Comment = entryComment
	}

	if entry.Type != "work" {
		entry.Start = normalizeTime(entry.Start)
		entry.End = normalizeTime(entry.End)
	}

	result, err := apiClient.UpdateTimeEntry(projectName, entry)
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
