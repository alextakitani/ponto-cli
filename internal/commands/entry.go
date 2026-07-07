package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Manage time entries",
	Long:  "Create, list, update, split, duplicate, and delete Ponto time entries.",
}

var entryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List time entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Get("/time_entries")
		if err != nil {
			return err
		}
		printList(enrichPresentation(resp.Data), entryColumns, fmt.Sprintf("%d entries", dataCount(resp.Data)), []Breadcrumb{
			breadcrumb("create", "ponto entry create --start \"2026-07-06 09:00\"", "Create a manual entry"),
		})
		return nil
	},
}

var entryShowCmd = &cobra.Command{
	Use:   "show ID",
	Short: "Show a time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Get("/time_entries/" + args[0])
		if err != nil {
			return err
		}
		printDetail(enrichPresentation(resp.Data), "Entry "+args[0], []Breadcrumb{
			breadcrumb("update", "ponto entry update "+args[0], "Update this entry"),
			breadcrumb("delete", "ponto entry delete "+args[0], "Delete this entry"),
		})
		return nil
	},
}

var entryDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		if _, err := c.Delete("/time_entries/" + args[0]); err != nil {
			return err
		}
		printMutation(nil, "Entry deleted", []Breadcrumb{
			breadcrumb("list", "ponto entry list", "List entries"),
		})
		return nil
	},
}

var entryCreateCmd = &cobra.Command{
	Use:   "create --start TS",
	Short: "Create a manual time entry",
	Long: `Create a manual time entry.

Timestamps accept RFC3339 or local forms like "2026-07-06 09:00"; the CLI sends an explicit local offset.
Omit --end to create a running timer. Omitted --billable lets the server derive billable from the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFlags(cmd, "start"); err != nil {
			return err
		}
		body, err := entryFlagBody(cmd, true)
		if err != nil {
			return err
		}
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Post("/time_entries", map[string]any{"time_entry": body})
		if err != nil {
			return handleTimerConflict(err)
		}
		printMutation(resp.Data, "Entry created", []Breadcrumb{
			breadcrumb("list", "ponto entry list", "List entries"),
		})
		return nil
	},
}

var entryUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a time entry",
	Long: `Update a time entry.

Only fields passed as flags are sent. The server silently ignores ended_at for a running timer; use ponto timer stop to stop it.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := entryFlagBody(cmd, false)
		if err != nil {
			return err
		}
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Patch("/time_entries/"+args[0], map[string]any{"time_entry": body})
		if err != nil {
			return err
		}
		printMutation(resp.Data, "Entry updated", []Breadcrumb{
			breadcrumb("show", "ponto entry show "+args[0], "Show this entry"),
		})
		return nil
	},
}

var entryDuplicateCmd = &cobra.Command{
	Use:   "duplicate ID",
	Short: "Duplicate an entry into a new timer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Post("/time_entries/"+args[0]+"/duplicate", nil)
		if err != nil {
			return handleTimerConflict(err)
		}
		printMutation(resp.Data, "Entry duplicated into running timer", []Breadcrumb{
			breadcrumb("status", "ponto timer status", "Show the running timer"),
		})
		return nil
	},
}

var entrySplitCmd = &cobra.Command{
	Use:   "split ID --at TS",
	Short: "Split a finished entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFlags(cmd, "at"); err != nil {
			return err
		}
		at, _ := cmd.Flags().GetString("at")
		cut, err := parseLocalTimestamp(at)
		if err != nil {
			return err
		}
		c, err := domainClient()
		if err != nil {
			return err
		}
		if _, err := c.Post("/time_entries/"+args[0]+"/split", map[string]any{"split": map[string]any{"cut": cut}}); err != nil {
			return err
		}
		printMutation(nil, "Entry split", []Breadcrumb{
			breadcrumb("list", "ponto entry list", "List entries"),
		})
		return nil
	},
}

func entryFlagBody(cmd *cobra.Command, create bool) (map[string]any, error) {
	body := map[string]any{}
	if changed(cmd, "start") {
		v, _ := cmd.Flags().GetString("start")
		ts, err := parseLocalTimestamp(v)
		if err != nil {
			return nil, err
		}
		body["started_at"] = ts
	} else if create {
		return nil, fmt.Errorf("--start is required")
	}
	if changed(cmd, "end") {
		v, _ := cmd.Flags().GetString("end")
		ts, err := parseLocalTimestamp(v)
		if err != nil {
			return nil, err
		}
		body["ended_at"] = ts
	}
	if projectID, ok, err := optionalInt(cmd, "project"); err != nil {
		return nil, err
	} else if ok {
		body["project_id"] = projectID
	}
	if taskID, ok, err := optionalInt(cmd, "task"); err != nil {
		return nil, err
	} else if ok {
		body["task_id"] = taskID
	}
	addStringIfChanged(cmd, body, "description", "description")
	if changed(cmd, "billable") && changed(cmd, "not-billable") {
		return nil, fmt.Errorf("--billable and --not-billable cannot be used together")
	}
	if changed(cmd, "billable") {
		body["billable"] = true
	}
	if changed(cmd, "not-billable") {
		body["billable"] = false
	}
	if changed(cmd, "tag") {
		tags, _ := cmd.Flags().GetIntSlice("tag")
		body["tag_ids"] = tags
	}
	if changed(cmd, "new-tag") {
		tags, _ := cmd.Flags().GetStringSlice("new-tag")
		body["new_tag_names"] = tags
	}
	return body, nil
}

func addEntryWriteFlags(cmd *cobra.Command) {
	cmd.Flags().String("start", "", "Start timestamp (RFC3339 or YYYY-MM-DD HH:MM)")
	cmd.Flags().String("end", "", "End timestamp (RFC3339 or YYYY-MM-DD HH:MM)")
	cmd.Flags().Int("project", 0, "Project ID; use 0 for no project")
	cmd.Flags().Int("task", 0, "Task ID")
	cmd.Flags().String("description", "", "Entry description")
	cmd.Flags().Bool("billable", false, "Mark entry billable")
	cmd.Flags().Bool("not-billable", false, "Mark entry not billable")
	cmd.Flags().IntSlice("tag", nil, "Tag ID (repeatable)")
	cmd.Flags().StringSlice("new-tag", nil, "Create or attach tag by name (repeatable)")
}

func init() {
	rootCmd.AddCommand(entryCmd)
	entryCmd.AddCommand(entryListCmd, entryShowCmd, entryCreateCmd, entryUpdateCmd, entryDeleteCmd, entryDuplicateCmd, entrySplitCmd)
	addEntryWriteFlags(entryCreateCmd)
	addEntryWriteFlags(entryUpdateCmd)
	entrySplitCmd.Flags().String("at", "", "Split timestamp (RFC3339 or YYYY-MM-DD HH:MM)")
}
