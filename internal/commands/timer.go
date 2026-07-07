package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var timerCmd = &cobra.Command{
	Use:   "timer",
	Short: "Track the running timer",
	Long:  "Start, stop, and inspect the singular running Ponto timer.",
}

var timerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a timer",
	Long: `Start the singular running timer.

Omit --project to let the server apply your active default project.
Use --project 0 or --no-project to explicitly start without a project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		if noProject, _ := cmd.Flags().GetBool("no-project"); noProject && changed(cmd, "project") {
			return fmt.Errorf("--project and --no-project cannot be used together")
		}
		timer := map[string]any{}
		if noProject, _ := cmd.Flags().GetBool("no-project"); noProject {
			timer["project_id"] = nil
		} else if projectID, ok, err := optionalInt(cmd, "project"); err != nil {
			return err
		} else if ok {
			timer["project_id"] = projectID
		}
		if taskID, ok, err := optionalInt(cmd, "task"); err != nil {
			return err
		} else if ok {
			timer["task_id"] = taskID
		}
		addStringIfChanged(cmd, timer, "description", "description")
		resp, err := c.Post("/timer", map[string]any{"timer": timer})
		if err != nil {
			return handleTimerConflict(err)
		}
		printMutation(resp.Data, "Timer started", []Breadcrumb{
			breadcrumb("status", "ponto timer status", "Show the running timer"),
			breadcrumb("stop", "ponto timer stop", "Stop the timer"),
		})
		return nil
	},
}

var timerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running timer",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Delete("/timer")
		if err != nil {
			return handleNoTimer(err)
		}
		if resp.StatusCode == 204 {
			printMutation(nil, "Timer discarded (duration 0:00:00)", []Breadcrumb{
				breadcrumb("start", "ponto timer start", "Start another timer"),
			})
			return nil
		}
		printMutation(resp.Data, "Timer stopped", []Breadcrumb{
			breadcrumb("list", "ponto entry list", "List time entries"),
			breadcrumb("start", "ponto timer start", "Start another timer"),
		})
		return nil
	},
}

var timerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the running timer",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Get("/timer")
		if err != nil {
			return err
		}
		if resp.Data == nil {
			printDetail(nil, "No timer running", []Breadcrumb{
				breadcrumb("start", "ponto timer start", "Start a timer"),
			})
			return nil
		}
		data := enrichPresentation(resp.Data)
		if m, ok := data.(map[string]any); ok {
			if started, _ := m["started_at"].(string); started != "" {
				if t, err := time.Parse(time.RFC3339Nano, started); err == nil {
					m["elapsed"] = formatDurationSeconds(int64(time.Since(t).Seconds()))
				}
			}
		}
		printDetail(data, "Timer running", []Breadcrumb{
			breadcrumb("stop", "ponto timer stop", "Stop the timer"),
			breadcrumb("entries", "ponto entry list", "List entries"),
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(timerCmd)
	timerCmd.AddCommand(timerStartCmd, timerStopCmd, timerStatusCmd)
	timerStartCmd.Flags().Int("project", 0, "Project ID; use 0 for no project")
	timerStartCmd.Flags().Bool("no-project", false, "Start explicitly without a project")
	timerStartCmd.Flags().Int("task", 0, "Task ID")
	timerStartCmd.Flags().String("description", "", "Timer description")
}
