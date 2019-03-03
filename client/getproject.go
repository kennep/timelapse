package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getProjectCmd)
}

var getProjectCmd = &cobra.Command{
	Use:   "get-project PROJECTNAME",
	Short: "Show a project",
	Long:  `Show some information about a project`,
	Args:  cobra.ExactArgs(1),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return getProject(args[0])
	}),
}

func getProject(projectName string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	project, err := apiClient.GetProject(projectName)
	if err != nil {
		return err
	}
	fmt.Println(project)

	return nil
}
