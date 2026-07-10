package commands

import (
	"net/url"
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
		if err := checkLimitAll(fetchAll(cmd)); err != nil {
			return err
		}
		projectID, _ := cmd.Flags().GetInt("project")
		c, err := domainClient()
		if err != nil {
			return err
		}
		values := url.Values{}
		applyPaginationParams(cmd, values)
		path := queryPath("/projects/"+strconv.Itoa(projectID)+"/tasks", values)
		resp, err := c.GetWithPagination(path, fetchAll(cmd))
		if err != nil {
			return err
		}
		printCollection(resp, resp.Data, taskColumns, "tasks", fetchAll(cmd), nil)
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
	addPaginationFlags(taskListCmd)
	taskCreateCmd.Flags().Int("project", 0, "Project ID")
	taskCreateCmd.Flags().String("name", "", "Task name")
	taskUpdateCmd.Flags().String("name", "", "Task name")
	taskCmd.AddCommand(
		taskListCmd,
		taskShowCmd,
		taskCreateCmd,
		taskUpdateCmd,
		taskDeleteCmd,
		newArchivalCmd("task", "tasks", taskArchiveBreadcrumbs),
		newUnarchivalCmd("task", "tasks", taskListBreadcrumbs),
	)
	rootCmd.AddCommand(taskCmd)
}

func taskArchiveBreadcrumbs(resource, plural, id string, data any) []Breadcrumb {
	breadcrumbs := []Breadcrumb{
		breadcrumb("show", "ponto "+resource+" show "+id, "Show this "+resource),
	}
	return append(breadcrumbs, taskListBreadcrumbs(resource, plural, id, data)...)
}

func taskListBreadcrumbs(_ string, _ string, _ string, data any) []Breadcrumb {
	cmd := "ponto task list"
	description := "List tasks"
	if projectID := projectIDFromTask(data); projectID != "" {
		cmd += " --project " + projectID
		description = "List project tasks"
	}
	return []Breadcrumb{
		breadcrumb("list", cmd, description),
	}
}

func projectIDFromTask(data any) string {
	task, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	projectID, ok := intFromAny(task["project_id"])
	if !ok {
		return ""
	}
	return strconv.FormatInt(projectID, 10)
}
