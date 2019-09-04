package client

import (
	"fmt"
	"os"

	"github.com/kennep/timelapse/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "timelapse",
	Short: "Timelapse is a project time tracker",
	Long:  `Timelapse is a project time tracker for the command line`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if logLevelTrace {
			logrus.Warn("Activating trace logging - this can cause confidential information to be logged!")
			logrus.SetLevel(logrus.TraceLevel)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var commandLineProject api.Project
var commandLineEntry api.TimeEntry

var logLevelTrace bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&logLevelTrace, "trace", false, "Activate trace logging")

}

// Execute is the main entry point to this command-line application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func _runCommand(cmd *cobra.Command, args []string, commandFunc func(*cobra.Command, []string) error) {
	err := commandFunc(cmd, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCommand(commandFunc func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		_runCommand(cmd, args, commandFunc)
	}
}
