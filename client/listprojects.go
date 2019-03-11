package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listProjectCmd)
}

var listProjectCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "Show a project",
	Long:  `Show some information about a project`,
	Args:  cobra.NoArgs,
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return listProjects()
	}),
}

func listProjects() error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	projects, err := apiClient.ListProjects()
	if err != nil {
		return err
	}
	for _, project := range projects {
		fmt.Println(project)
	}

	return nil
}
