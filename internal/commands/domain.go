package commands

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alextakitani/ponto-cli/internal/client"
	pontoerrors "github.com/alextakitani/ponto-cli/internal/errors"
	"github.com/basecamp/cli/output"
	"github.com/spf13/cobra"
)

func domainClient() (clientAPI, error) {
	if err := requireAPIConfig(); err != nil {
		return nil, err
	}
	return getClient(), nil
}

type clientAPI interface {
	Get(path string) (*client.APIResponse, error)
	Post(path string, body any) (*client.APIResponse, error)
	Patch(path string, body any) (*client.APIResponse, error)
	Delete(path string) (*client.APIResponse, error)
}

func parseLocalTimestamp(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("timestamp is required")
	}
	offsetLayouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04-07:00",
	}
	for _, layout := range offsetLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.Format(time.RFC3339), nil
		}
	}
	localLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range localLayouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t.Format(time.RFC3339), nil
		}
	}
	return "", fmt.Errorf("invalid timestamp %q; use RFC3339 or YYYY-MM-DD HH:MM", value)
}

func formatDurationSeconds(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}

func formatMoney(rateCents any, currency any) string {
	c, ok := intFromAny(rateCents)
	if !ok {
		return ""
	}
	cur, _ := currency.(string)
	if cur == "" {
		return fmt.Sprintf("%.2f", float64(c)/100)
	}
	return fmt.Sprintf("%.2f %s", float64(c)/100, cur)
}

func intFromAny(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

func enrichPresentation(data any) any {
	switch v := data.(type) {
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = enrichPresentation(item)
		}
		return out
	case map[string]any:
		m := copyMap(v)
		if _, ok := m["duration"]; !ok {
			if sec, ok := intFromAny(m["duration_seconds"]); ok {
				m["duration"] = formatDurationSeconds(sec)
			} else if _, running := m["ended_at"]; running && m["ended_at"] == nil {
				m["duration"] = "running"
			}
		}
		if _, ok := m["rate"]; !ok {
			if rate := formatMoney(m["rate_cents"], m["currency"]); rate != "" {
				m["rate"] = rate
			}
		}
		if _, ok := m["effective_rate"]; !ok {
			if rate := formatMoney(m["effective_rate_cents"], m["effective_currency"]); rate != "" {
				m["effective_rate"] = rate
			}
		}
		return m
	default:
		return data
	}
}

func copyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func queryPath(base string, values url.Values) string {
	if len(values) == 0 {
		return base
	}
	return base + "?" + values.Encode()
}

func addCommonListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("archived", false, "List archived records")
	cmd.Flags().String("query", "", "Search by name")
}

func applyCommonListParams(cmd *cobra.Command, values url.Values) {
	if archived, _ := cmd.Flags().GetBool("archived"); archived {
		values.Set("archived", "1")
	}
	if q, _ := cmd.Flags().GetString("query"); strings.TrimSpace(q) != "" {
		values.Set("q", strings.TrimSpace(q))
	}
}

func changed(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Changed(name)
}

func optionalInt(cmd *cobra.Command, name string) (any, bool, error) {
	if !changed(cmd, name) {
		return nil, false, nil
	}
	v, err := cmd.Flags().GetInt(name)
	if err != nil {
		return nil, false, err
	}
	if v == 0 {
		return nil, true, nil
	}
	return v, true, nil
}

func addStringIfChanged(cmd *cobra.Command, body map[string]any, flag, key string) {
	if changed(cmd, flag) {
		v, _ := cmd.Flags().GetString(flag)
		body[key] = v
	}
}

func addIntIfChanged(cmd *cobra.Command, body map[string]any, flag, key string) {
	if changed(cmd, flag) {
		v, _ := cmd.Flags().GetInt(flag)
		body[key] = v
	}
}

func requireFlags(cmd *cobra.Command, names ...string) error {
	for _, name := range names {
		if !cmd.Flags().Changed(name) {
			return pontoerrors.NewInvalidArgsError("--" + name + " is required")
		}
	}
	return nil
}

func handleTimerConflict(err error) error {
	var e *output.Error
	if errors.As(err, &e) && e.HTTPStatus == 409 {
		return &output.Error{
			Code:       e.Code,
			Message:    "timer is already running",
			HTTPStatus: e.HTTPStatus,
			Hint:       "Run 'ponto timer status' or 'ponto timer stop'",
		}
	}
	return err
}

func handleNoTimer(err error) error {
	var e *output.Error
	if errors.As(err, &e) && e.HTTPStatus == 404 {
		return pontoerrors.NewNotFoundError("no timer running")
	}
	return err
}
