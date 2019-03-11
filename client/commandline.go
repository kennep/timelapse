package client

import (
	"fmt"
	"os"

	"github.com/kennep/timelapse/api"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "timelapse",
	Short: "Timelapse is a project time tracker",
	Long:  `Timelapse is a project time tracker for the command line`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var commandLineProject api.Project

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
