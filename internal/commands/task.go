package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

var taskListCmd = &cobra.Command{
	Use:   "list --project ID",
	Short: "List tasks for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFlags(cmd, "project"); err != nil {
			return err
		}
		projectID, _ := cmd.Flags().GetInt("project")
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Get("/projects/" + strconv.Itoa(projectID) + "/tasks")
		if err != nil {
			return err
		}
		printList(resp.Data, taskColumns, fmt.Sprintf("%d tasks", dataCount(resp.Data)), nil)
		return nil
	},
}

var taskShowCmd = &cobra.Command{
	Use:   "show ID",
	Short: "Show a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Get("/tasks/" + args[0])
		if err != nil {
			return err
		}
		printDetail(resp.Data, "Task "+args[0], nil)
		return nil
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create --project ID --name NAME",
	Short: "Create a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFlags(cmd, "project", "name"); err != nil {
			return err
		}
		projectID, _ := cmd.Flags().GetInt("project")
		name, _ := cmd.Flags().GetString("name")
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Post("/projects/"+strconv.Itoa(projectID)+"/tasks", map[string]any{"task": map[string]any{"name": name}})
		if err != nil {
			return err
		}
		printMutation(resp.Data, "Task created", []Breadcrumb{
			breadcrumb("list", "ponto task list --project "+strconv.Itoa(projectID), "List project tasks"),
		})
		return nil
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		addStringIfChanged(cmd, body, "name", "name")
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Patch("/tasks/"+args[0], map[string]any{"task": body})
		if err != nil {
			return err
		}
		printMutation(resp.Data, "Task updated", nil)
		return nil
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		if _, err := c.Delete("/tasks/" + args[0]); err != nil {
			return err
		}
		printMutation(nil, "Task deleted", nil)
		return nil
	},
}

func init() {
	taskListCmd.Flags().Int("project", 0, "Project ID")
	taskCreateCmd.Flags().Int("project", 0, "Project ID")
	taskCreateCmd.Flags().String("name", "", "Task name")
	taskUpdateCmd.Flags().String("name", "", "Task name")
	taskCmd.AddCommand(taskListCmd, taskShowCmd, taskCreateCmd, taskUpdateCmd, taskDeleteCmd)
	rootCmd.AddCommand(taskCmd)
}
