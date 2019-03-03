package client

import (
	"fmt"

	"github.com/kennep/timelapse/domain"
	"github.com/spf13/cobra"
)

var updateProjectOptions domain.Project

func init() {
	rootCmd.AddCommand(updateProjectCmd)
	updateProjectCmd.Flags().StringVarP(&updateProjectOptions.Name, "rename-to", "r", "", "Rename project")
	updateProjectCmd.Flags().StringVarP(&updateProjectOptions.Description, "description", "d", "", "Set description")
	updateProjectCmd.Flags().BoolVarP(&updateProjectOptions.Billable, "billable", "b", false, "Mark a project as billable")
}

var updateProjectCmd = &cobra.Command{
	Use:   "update-project PROJECTNAME",
	Short: "Update a project",
	Long:  `Update the specified project`,
	Args:  cobra.RangeArgs(1, 2),
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return updateProject(cmd, args[0])
	}),
}

func updateProject(cmd *cobra.Command, projectName string) error {
	apiClient, err := NewApiClient()
	if err != nil {
		return err
	}

	var updateProjectRequest UpdateProjectRequest
	if cmd.Flags().Changed("rename-to") {
		updateProjectRequest.Name = &updateProjectOptions.Name
	}
	if cmd.Flags().Changed("description") {
		updateProjectRequest.Description = &updateProjectOptions.Description
	}
	if cmd.Flags().Changed("billable") {
		updateProjectRequest.Billable = &updateProjectOptions.Billable
	}

	project, err := apiClient.UpdateProject(projectName, &updateProjectRequest)
	if err != nil {
		return err
	}
	fmt.Println(project)

	return nil
}
