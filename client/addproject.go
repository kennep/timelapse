package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addProjectCmd)
	addProjectCmd.Flags().BoolVarP(&commandLineProject.Billable, "billable", "b", false, "Mark a project as billable")
	addProjectCmd.Flags().StringVarP(&commandLineProject.Description, "description", "d", "", "Project description")
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

	commandLineProject.Name = projectName

	project, err := apiClient.CreateProject(&commandLineProject)
	if err != nil {
		return err
	}
	fmt.Println(project)

	return nil
}
