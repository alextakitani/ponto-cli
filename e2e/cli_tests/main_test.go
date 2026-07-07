package cli_tests

import (
	"testing"

	"github.com/alextakitani/ponto-cli/e2e/harness"
)

func TestCommandsCatalogIncludesDomainCommands(t *testing.T) {
	h := harness.New(t)
	result := h.Run("commands", "--json")
	requireOK(t, result)
	found := map[string]bool{}
	var walk func([]any)
	walk = func(items []any) {
		for _, item := range items {
			m, _ := item.(map[string]any)
			name, _ := m["name"].(string)
			if name != "" {
				found[name] = true
			}
			if children, ok := m["subcommands"].([]any); ok {
				walk(children)
			}
		}
	}
	walk(result.GetDataArray())
	for _, name := range []string{
		"ponto timer", "ponto timer start", "ponto timer status", "ponto timer stop",
		"ponto entry", "ponto client", "ponto project", "ponto task", "ponto tag",
		"ponto client archive", "ponto client unarchive", "ponto export",
	} {
		if !found[name] {
			t.Fatalf("commands catalog missing %s", name)
		}
	}
}

func requireOK(t *testing.T, result *harness.Result) {
	t.Helper()
	if result.ExitCode != 0 {
		t.Fatalf("exit=%d stdout=%s stderr=%s", result.ExitCode, result.Stdout, result.Stderr)
	}
	if result.ParseError != nil {
		t.Fatalf("parse response: %v stdout=%s", result.ParseError, result.Stdout)
	}
	if result.Response == nil || !result.Response.OK {
		t.Fatalf("response not ok: %#v stdout=%s", result.Response, result.Stdout)
	}
}
