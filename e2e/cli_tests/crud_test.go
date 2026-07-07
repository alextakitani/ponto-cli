package cli_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alextakitani/ponto-cli/e2e/harness"
)

func TestTimerStartStatusStop(t *testing.T) {
	h := harness.New(t)
	_ = h.Run("timer", "stop")

	result := h.Run("timer", "start", "--no-project", "--description", "ponto-cli e2e timer")
	requireOK(t, result)
	if result.Response.Summary != "Timer started" {
		t.Fatalf("summary=%q", result.Response.Summary)
	}

	status := h.Run("timer", "status")
	requireOK(t, status)
	if status.GetDataString("description") != "ponto-cli e2e timer" {
		t.Fatalf("unexpected timer status data: %#v", status.GetDataMap())
	}

	stop := h.Run("timer", "stop")
	requireOK(t, stop)
}

func TestCatalogAndEntryCRUD(t *testing.T) {
	h := harness.New(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	client := h.Run("client", "create", "--name", "CLI Client "+suffix, "--currency", "BRL", "--rate-cents", "15000")
	requireOK(t, client)
	clientID := client.GetDataInt("id")
	defer h.Run("client", "delete", fmt.Sprintf("%d", clientID))

	clientList := h.Run("client", "list", "--query", suffix)
	requireOK(t, clientList)
	clientShow := h.Run("client", "show", fmt.Sprintf("%d", clientID))
	requireOK(t, clientShow)
	clientUpdate := h.Run("client", "update", fmt.Sprintf("%d", clientID), "--note", "updated by e2e")
	requireOK(t, clientUpdate)
	clientArchive := h.Run("client", "archive", fmt.Sprintf("%d", clientID))
	requireOK(t, clientArchive)
	clientUnarchive := h.Run("client", "unarchive", fmt.Sprintf("%d", clientID))
	requireOK(t, clientUnarchive)

	project := h.Run("project", "create", "--name", "CLI Project "+suffix, "--client", fmt.Sprintf("%d", clientID), "--color", "#1e66f5")
	requireOK(t, project)
	projectID := project.GetDataInt("id")
	defer h.Run("project", "delete", fmt.Sprintf("%d", projectID))

	projectList := h.Run("project", "list", "--client", fmt.Sprintf("%d", clientID), "--query", suffix)
	requireOK(t, projectList)
	projectShow := h.Run("project", "show", fmt.Sprintf("%d", projectID))
	requireOK(t, projectShow)
	projectUpdate := h.Run("project", "update", fmt.Sprintf("%d", projectID), "--rate-cents", "20000")
	requireOK(t, projectUpdate)
	projectDefault := h.Run("project", "default", fmt.Sprintf("%d", projectID))
	requireOK(t, projectDefault)
	projectDefaultClear := h.Run("project", "default", "--clear")
	requireOK(t, projectDefaultClear)

	task := h.Run("task", "create", "--project", fmt.Sprintf("%d", projectID), "--name", "CLI Task "+suffix)
	requireOK(t, task)
	taskID := task.GetDataInt("id")
	defer h.Run("task", "delete", fmt.Sprintf("%d", taskID))

	taskList := h.Run("task", "list", "--project", fmt.Sprintf("%d", projectID))
	requireOK(t, taskList)
	taskShow := h.Run("task", "show", fmt.Sprintf("%d", taskID))
	requireOK(t, taskShow)
	taskUpdate := h.Run("task", "update", fmt.Sprintf("%d", taskID), "--name", "CLI Task Updated "+suffix)
	requireOK(t, taskUpdate)

	tag := h.Run("tag", "create", "--name", "cli-tag-"+suffix)
	requireOK(t, tag)
	tagID := tag.GetDataInt("id")
	defer h.Run("tag", "delete", fmt.Sprintf("%d", tagID))

	tagList := h.Run("tag", "list", "--query", suffix)
	requireOK(t, tagList)
	tagShow := h.Run("tag", "show", fmt.Sprintf("%d", tagID))
	requireOK(t, tagShow)
	tagUpdate := h.Run("tag", "update", fmt.Sprintf("%d", tagID), "--name", "cli-tag-updated-"+suffix)
	requireOK(t, tagUpdate)

	start := time.Now().Add(-72 * time.Hour).Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	entry := h.Run("entry", "create",
		"--start", start.Format(time.RFC3339),
		"--end", end.Format(time.RFC3339),
		"--project", fmt.Sprintf("%d", projectID),
		"--task", fmt.Sprintf("%d", taskID),
		"--description", "CLI Entry "+suffix,
		"--billable",
		"--tag", fmt.Sprintf("%d", tagID),
		"--new-tag", "cli-inline-"+suffix,
	)
	requireOK(t, entry)
	entryID := entry.GetDataInt("id")
	defer h.Run("entry", "delete", fmt.Sprintf("%d", entryID))

	entryList := h.Run("entry", "list")
	requireOK(t, entryList)
	entryShow := h.Run("entry", "show", fmt.Sprintf("%d", entryID))
	requireOK(t, entryShow)
	entryUpdate := h.Run("entry", "update", fmt.Sprintf("%d", entryID), "--description", "CLI Entry Updated "+suffix, "--not-billable")
	requireOK(t, entryUpdate)
	splitAt := start.Add(15 * time.Minute)
	entrySplit := h.Run("entry", "split", fmt.Sprintf("%d", entryID), "--at", splitAt.Format(time.RFC3339))
	requireOK(t, entrySplit)
}

func TestReportExportCSV(t *testing.T) {
	h := harness.New(t)
	path := filepath.Join(t.TempDir(), "ponto-report.csv")
	result := h.Run("export", "--format", "csv", "--period", "month", "--output", path)
	requireOK(t, result)
	if result.Response.Summary != "Report exported to "+path {
		t.Fatalf("summary=%q", result.Response.Summary)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("export file missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("export file is empty")
	}
}
