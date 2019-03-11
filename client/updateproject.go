package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateProjectCmd)
	updateProjectCmd.Flags().StringVarP(&commandLineProject.Name, "rename-to", "r", "", "Rename project")
	updateProjectCmd.Flags().StringVarP(&commandLineProject.Description, "description", "d", "", "Set description")
	updateProjectCmd.Flags().BoolVarP(&commandLineProject.Billable, "billable", "b", false, "Mark a project as billable")
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

	project, err := apiClient.GetProject(projectName)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("rename-to") {
		project.Name = commandLineProject.Name
	}
	if cmd.Flags().Changed("description") {
		project.Description = commandLineProject.Description
	}
	if cmd.Flags().Changed("billable") {
		project.Billable = commandLineProject.Billable
	}

	project, err = apiClient.UpdateProject(projectName, project)
	if err != nil {
		return err
	}
	fmt.Println(project)

	return nil
}
