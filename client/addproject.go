package client

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kennep/timelapse/domain"
)

var addProjectOptions domain.Project

func init() {
	rootCmd.AddCommand(addProjectCmd)
	addProjectCmd.Flags().BoolVarP(&addProjectOptions.Billable, "billable", "b", false, "Mark a project as billable")
	addProjectCmd.Flags().StringVarP(&addProjectOptions.Description, "description", "d", "", "Project description")
}

var addProjectCmd = &cobra.Command{
	Use:   "add-project PROJECTNAME",
	Short: "Add a project",
	Long:  `Add the specified project`,
	Args:  cobra.ExactArgs(1),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return addProject(args[0])
	}),
}

func addProject(projectName string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	addProjectOptions.Name = projectName

	project, err := apiClient.CreateProject(&addProjectOptions)
	if err != nil {
		return err
	}
	fmt.Println(project)

	return nil
}
