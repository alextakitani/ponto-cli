package errors

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/basecamp/cli/output"
)

// The READMEs publish a table pairing every error `code` with the exit status
// it produces. Integrators branch on those numbers, so a table that drifts from
// ExitCodeFor is worse than no table at all — it is a promise the binary stops
// keeping, silently. This reads the tables back and checks every row.
//
// Rows look like: |  `3` | `auth_required` | Not authenticated ... |
var exitRow = regexp.MustCompile("(?m)^\\|\\s*`(\\d+)`\\s*\\|\\s*`([a-z_]+)`\\s*\\|")

func TestREADMEExitCodeTablesMatchImplementation(t *testing.T) {
	for _, name := range []string{"README.md", "README.pt.md"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", name)
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}

			rows := exitRow.FindAllStringSubmatch(string(body), -1)
			if len(rows) == 0 {
				t.Fatalf("%s: no exit-code table found; the documented contract went missing", name)
			}

			seen := make(map[string]bool, len(rows))
			for _, row := range rows {
				documented, err := strconv.Atoi(row[1])
				if err != nil {
					t.Fatalf("%s: unparseable exit status %q", name, row[1])
				}
				code := row[2]
				seen[code] = true

				if actual := output.ExitCodeFor(code); actual != documented {
					t.Errorf("%s documents code %q as exit %d, but ExitCodeFor returns %d",
						name, code, documented, actual)
				}
			}

			// Every code the binary can emit must appear, or a caller reading
			// the table would have no branch for an outcome it will really hit.
			for _, code := range []string{
				output.CodeUsage, output.CodeNotFound, output.CodeAuth,
				output.CodeForbidden, output.CodeRateLimit, output.CodeNetwork,
				output.CodeAPI, output.CodeAmbiguous,
			} {
				if !seen[code] {
					t.Errorf("%s: code %q is missing from the exit-code table", name, code)
				}
			}
		})
	}
}

// The READMEs also promise that an unrecognised code degrades to ExitAPI rather
// than passing as success — the property that keeps a future code from being
// read as "it worked".
func TestUnknownCodeDegradesToAPIError(t *testing.T) {
	for _, code := range []string{"", "some_future_code", "TOTALLY_UNKNOWN"} {
		if got := output.ExitCodeFor(code); got != ExitAPI {
			t.Errorf("ExitCodeFor(%q) = %d, want %d (documented as never 0)", code, got, ExitAPI)
		}
		if output.ExitCodeFor(code) == ExitSuccess {
			t.Errorf("ExitCodeFor(%q) returned success for an error code", code)
		}
	}
}
