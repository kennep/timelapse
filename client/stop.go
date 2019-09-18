package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/kennep/timelapse/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringVarP(&breaks, "breaks", "b", "", "Break duration")
	stopCmd.Flags().StringVarP(&entryType, "type", "t", "work", "Entry type (work|sick|sick-child|vacation)")
	stopCmd.Flags().StringVarP(&entryComment, "comment", "c", "", "Entry comment")
}

var stopCmd = &cobra.Command{
	Use:   "stop [PROJECTNAME] [END]",
	Short: "Stop time tracking",
	Long:  `stop time tracking for a given project`,
	Args:  cobra.RangeArgs(0, 2),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return stopTimeEntry(args[0], args[1])
		}
		if len(args) > 0 {
			return stopTimeEntry(args[0], "")
		}
		return stopTimeEntry("", "")
	}),
}

func stopTimeEntry(projectName string, endTime string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	entries, err := apiClient.GetTimeEntries(projectName)
	if err != nil {
		return err
	}

	var foundEntry *api.TimeEntry

	for _, entry := range entries {
		if entry.End == nil {
			foundEntry = entry
		}
	}
	if foundEntry == nil {
		return errors.New("Did not find any time entry to close")
	}

	now := time.Now()

	foundEntry.End = &now

	if endTime != "" {
		tm, err := ParseTimeRef(endTime, now)
		foundEntry.End = &tm
		if err != nil {
			return err
		}
	}

	if breaks != "" {
		foundEntry.Breaks, err = time.ParseDuration(breaks)
		if err != nil {
			return err
		}
	}

	if entryType != "work" {
		*foundEntry.End = normalizeTime(*foundEntry.End)
	}

	result, err := apiClient.UpdateTimeEntry(foundEntry.ProjectName, foundEntry)
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
