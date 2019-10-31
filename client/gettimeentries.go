package client

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getTimeEntriesCmd)
	getTimeEntriesCmd.Flags().StringVarP(&startTime, "start", "s", "", "Only entries that start later than this start time/date")
	getTimeEntriesCmd.Flags().StringVarP(&endTime, "end", "e", "", "Only entries that end earlier than this time/date")
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

	var start time.Time
	var end time.Time
	now := time.Now()
	if startTime != "" {
		start, err = ParseTimeRef(startTime, now)
		if err != nil {
			return err
		}
	}

	if endTime != "" {
		end, err = ParseTimeRef(endTime, now)
		if err != nil {
			return err
		}
	}

	var totalDuration time.Duration
	for _, timeentry := range timeentries {
		if !start.IsZero() && (timeentry.Start == nil || timeentry.Start.Before(start)) {
			continue
		}
		if !end.IsZero() && (timeentry.End == nil || timeentry.End.After(end)) {
			continue
		}
		fmt.Println(timeentry)
		totalDuration += timeentry.Duration()
	}
	fmt.Println("Total duration: " + totalDuration.String())

	return nil
}
