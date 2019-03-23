package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/kennep/timelapse/api"
	"github.com/spf13/cobra"
)

var startTime string
var endTime string
var breaks string
var entryType string
var entryComment string

func init() {
	rootCmd.AddCommand(addTimeEntryCmd)
	addTimeEntryCmd.Flags().StringVarP(&startTime, "start", "s", "", "Entry start time/date")
	addTimeEntryCmd.Flags().StringVarP(&endTime, "end", "e", "", "Entry end time/date")
	addTimeEntryCmd.Flags().StringVarP(&breaks, "breaks", "b", "", "Break duration")
	addTimeEntryCmd.Flags().StringVarP(&entryType, "type", "t", "work", "Entry type (work|sick|sick-child|vacation)")
	addTimeEntryCmd.Flags().StringVarP(&entryComment, "comment", "c", "", "Entry comment")
}

var addTimeEntryCmd = &cobra.Command{
	Use:   "add-entry PROJECTNAME [-s START] [-e END]",
	Short: "Add time entry",
	Long:  `Add a time entry for a given project`,
	Args:  cobra.ExactArgs(1),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return addTimeEntry(args[0])
	}),
}

func addTimeEntry(projectName string) error {
	if startTime == "" && endTime == "" {
		return errors.New("At least one of start time or end time must be specified!")
	}

	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	entry := api.TimeEntry{
		Start:   time.Now(),
		End:     time.Now(),
		Breaks:  time.Duration(0),
		Type:    entryType,
		Comment: entryComment,
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

	if breaks != "" {
		entry.Breaks, err = time.ParseDuration(breaks)
		if err != nil {
			return err
		}
	}

	if entryType != "work" {
		entry.Start = zeroTime(entry.Start)
		entry.End = zeroTime(entry.End)
	}

	result, err := apiClient.AddTimeEntry(projectName, &entry)
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}

func zeroTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
