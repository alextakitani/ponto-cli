package commands

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestGenerateSurfaceSnapshot(t *testing.T) {
	if os.Getenv("GENERATE_SURFACE") == "" {
		t.Skip("set GENERATE_SURFACE=1 to regenerate SURFACE.txt")
	}
	writeSurfaceSnapshot(t)
}

func TestSurfaceSnapshot(t *testing.T) {
	got := surfaceSnapshot()
	path := filepath.Join("..", "..", "SURFACE.txt")
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(want)) != strings.TrimSpace(got) {
		t.Fatalf("SURFACE.txt is stale; run make surface-snapshot")
	}
}

func writeSurfaceSnapshot(t *testing.T) {
	t.Helper()
	path := filepath.Join("..", "..", "SURFACE.txt")
	if err := os.WriteFile(path, []byte(surfaceSnapshot()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func surfaceSnapshot() string {
	configureCLIUX()
	seen := map[string]bool{}
	var lines []string
	add := func(line string) {
		if seen[line] {
			return
		}
		seen[line] = true
		lines = append(lines, line)
	}
	walkCommandTree(rootCmd, func(cmd *cobra.Command) {
		if cmd.Hidden {
			return
		}
		add("CMD " + cmd.CommandPath())
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if !f.Hidden {
				add(surfaceFlagLine(cmd, f))
			}
		})
		if cmd == rootCmd {
			cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				if !f.Hidden {
					add(surfaceFlagLine(cmd, f))
				}
			})
		}
	})
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func surfaceFlagLine(cmd *cobra.Command, f *pflag.Flag) string {
	line := "FLAG " + cmd.CommandPath() + " --" + f.Name + " type=" + f.Value.Type()
	if f.Shorthand != "" {
		line += " shorthand=" + f.Shorthand
	}
	return line
}
