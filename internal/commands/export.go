package commands

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pontoerrors "github.com/alextakitani/ponto-cli/internal/errors"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a report file",
	Long: `Export finished time entries as CSV or XLSX.

By default the report is written to the filename returned by the server in the current directory. Use --output - to write raw file bytes to stdout without the JSON envelope.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := exportPath(cmd)
		if err != nil {
			return err
		}

		c, err := domainClient()
		if err != nil {
			return err
		}
		body, filename, err := c.DownloadFile(path)
		if err != nil {
			return err
		}

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "-" {
			w := outWriter
			if w == nil {
				w = os.Stdout
			}
			_, err := w.Write(body)
			return err
		}

		if outputPath == "" {
			outputPath = safeDownloadFilename(filename, exportFormat(cmd))
		}
		if err := os.WriteFile(outputPath, body, 0o644); err != nil {
			return pontoerrors.NewError(fmt.Sprintf("writing export: %v", err))
		}

		printMutation(map[string]any{
			"file":  outputPath,
			"bytes": len(body),
		}, "Report exported to "+outputPath, nil)
		return nil
	},
}

func init() {
	exportCmd.Flags().String("format", "csv", "Export format (csv or xlsx)")
	exportCmd.Flags().String("period", "month", "Report period (today, week, month, year, last_month, last_year, custom)")
	exportCmd.Flags().String("from", "", "Start date for --period custom (YYYY-MM-DD)")
	exportCmd.Flags().String("to", "", "End date for --period custom (YYYY-MM-DD)")
	exportCmd.Flags().StringArray("client", nil, "Client ID filter (repeatable; use 'none' for no client)")
	exportCmd.Flags().StringArray("project", nil, "Project ID filter (repeatable; use 'none' for no project)")
	exportCmd.Flags().StringArray("task", nil, "Task ID filter (repeatable; use 'none' for no task)")
	exportCmd.Flags().StringArray("tag", nil, "Tag ID filter (repeatable; use 'none' for no tag)")
	exportCmd.Flags().Bool("billable", false, "Filter billable entries")
	exportCmd.Flags().Bool("not-billable", false, "Filter non-billable entries")
	exportCmd.Flags().String("description", "", "Filter by description text")
	exportCmd.Flags().String("group-by", "", "Primary grouping (project, client, task, tag, description)")
	exportCmd.Flags().String("group-by-2", "", "Secondary grouping (project, client, task, tag, description)")
	exportCmd.Flags().Bool("rounding", false, "Enable report rounding")
	exportCmd.Flags().Int("rounding-block", 0, "Rounding block in minutes (5, 15, or 30)")
	exportCmd.Flags().String("rounding-direction", "", "Rounding direction (nearest, up, or down)")
	exportCmd.Flags().String("locale", "", "Export locale (pt-BR or en)")
	exportCmd.Flags().StringP("output", "o", "", "Output path; use '-' for raw stdout")
	rootCmd.AddCommand(exportCmd)
}

func exportPath(cmd *cobra.Command) (string, error) {
	format := exportFormat(cmd)
	if !oneOf(format, "csv", "xlsx") {
		return "", pontoerrors.NewInvalidArgsError("--format must be csv or xlsx")
	}

	values := url.Values{}
	if err := addEnumParam(cmd, values, "period", "period", "today", "week", "month", "year", "last_month", "last_year", "custom"); err != nil {
		return "", err
	}
	period, _ := cmd.Flags().GetString("period")
	if period == "custom" {
		if err := requireFlags(cmd, "from", "to"); err != nil {
			return "", err
		}
	}
	if err := addDateParam(cmd, values, "from", "from"); err != nil {
		return "", err
	}
	if err := addDateParam(cmd, values, "to", "to"); err != nil {
		return "", err
	}
	for _, filter := range []struct {
		flag string
		key  string
	}{
		{"client", "client_ids[]"},
		{"project", "project_ids[]"},
		{"task", "task_ids[]"},
		{"tag", "tag_ids[]"},
	} {
		if err := addIDFilters(cmd, values, filter.flag, filter.key); err != nil {
			return "", err
		}
	}
	if changed(cmd, "billable") && changed(cmd, "not-billable") {
		return "", pontoerrors.NewInvalidArgsError("--billable and --not-billable cannot be used together")
	}
	if changed(cmd, "billable") {
		values.Set("billable", "true")
	}
	if changed(cmd, "not-billable") {
		values.Set("billable", "false")
	}
	addTrimmedStringParam(cmd, values, "description", "description")
	if err := addEnumParam(cmd, values, "group-by", "group_by", "project", "client", "task", "tag", "description"); err != nil {
		return "", err
	}
	if err := addEnumParam(cmd, values, "group-by-2", "group_by_2", "project", "client", "task", "tag", "description"); err != nil {
		return "", err
	}
	if changed(cmd, "rounding") {
		rounding, _ := cmd.Flags().GetBool("rounding")
		if rounding {
			values.Set("rounding", "on")
		}
	}
	if changed(cmd, "rounding-block") {
		block, _ := cmd.Flags().GetInt("rounding-block")
		if block != 5 && block != 15 && block != 30 {
			return "", pontoerrors.NewInvalidArgsError("--rounding-block must be 5, 15, or 30")
		}
		values.Set("rounding_block", strconv.Itoa(block))
	}
	if err := addEnumParam(cmd, values, "rounding-direction", "rounding_direction", "nearest", "up", "down"); err != nil {
		return "", err
	}
	if err := addEnumParam(cmd, values, "locale", "export_locale", "pt-BR", "en"); err != nil {
		return "", err
	}
	return queryPath("/reports/export."+format, values), nil
}

func exportFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("format")
	return strings.ToLower(strings.TrimSpace(format))
}

func addEnumParam(cmd *cobra.Command, values url.Values, flag, key string, allowed ...string) error {
	if !changed(cmd, flag) {
		return nil
	}
	value, _ := cmd.Flags().GetString(flag)
	value = strings.TrimSpace(value)
	if !oneOf(value, allowed...) {
		return pontoerrors.NewInvalidArgsError("--" + flag + " must be one of: " + strings.Join(allowed, ", "))
	}
	values.Set(key, value)
	return nil
}

func addDateParam(cmd *cobra.Command, values url.Values, flag, key string) error {
	if !changed(cmd, flag) {
		return nil
	}
	value, _ := cmd.Flags().GetString(flag)
	value = strings.TrimSpace(value)
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return pontoerrors.NewInvalidArgsError("--" + flag + " must use YYYY-MM-DD")
	}
	values.Set(key, value)
	return nil
}

func addIDFilters(cmd *cobra.Command, values url.Values, flag, key string) error {
	if !changed(cmd, flag) {
		return nil
	}
	items, _ := cmd.Flags().GetStringArray(flag)
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if item != "none" {
			if _, err := strconv.ParseInt(item, 10, 64); err != nil {
				return pontoerrors.NewInvalidArgsError("--" + flag + " must be an ID or 'none'")
			}
		}
		values.Add(key, item)
	}
	return nil
}

func addTrimmedStringParam(cmd *cobra.Command, values url.Values, flag, key string) {
	if !changed(cmd, flag) {
		return
	}
	value, _ := cmd.Flags().GetString(flag)
	if value = strings.TrimSpace(value); value != "" {
		values.Set(key, value)
	}
}

func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func safeDownloadFilename(filename, format string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "ponto-report." + format
	}
	filename = filepath.Base(filename)
	if filename == "." || filename == string(filepath.Separator) {
		return "ponto-report." + format
	}
	return filename
}
