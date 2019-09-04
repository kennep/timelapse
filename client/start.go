package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/kennep/timelapse/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&breaks, "breaks", "b", "", "Break duration")
	startCmd.Flags().StringVarP(&entryType, "type", "t", "work", "Entry type (work|sick|sick-child|vacation)")
	startCmd.Flags().StringVarP(&entryComment, "comment", "c", "", "Entry comment")
}

var startCmd = &cobra.Command{
	Use:   "start PROJECTNAME [START]",
	Short: "Start time tracking",
	Long:  `start time tracking for a given project`,
	Args:  cobra.RangeArgs(1, 2),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return startTimeEntry(args[0], args[1])
		}
		return startTimeEntry(args[0], "")
	}),
}

func startTimeEntry(projectName string, startTime string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	if projectName == "" {
		return errors.New("Project name must be given")
	}

	entries, err := apiClient.GetTimeEntries(projectName)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.End == nil {
			return errors.New("Already started entry: \"" + entry.String() + "\" - close it first")
		}
	}

	now := time.Now()

	entry := api.TimeEntry{
		Start:   &now,
		End:     nil,
		Breaks:  time.Duration(0),
		Type:    entryType,
		Comment: entryComment,
	}

	if startTime != "" {
		*entry.Start, err = ParseTimeRef(startTime, *entry.Start)
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
		*entry.Start = normalizeTime(*entry.Start)
	}

	result, err := apiClient.AddTimeEntry(projectName, &entry)
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
