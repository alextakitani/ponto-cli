package commands

import (
	"strings"
	"testing"
	"time"
)

func TestParseLocalTimestampRFC3339(t *testing.T) {
	got, err := parseLocalTimestamp("2026-07-06T09:00:00-03:00")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2026-07-06T09:00:00-03:00" {
		t.Fatalf("got %q", got)
	}
}

func TestParseLocalTimestampLocalAddsOffset(t *testing.T) {
	old := time.Local
	t.Cleanup(func() { time.Local = old })
	loc := time.FixedZone("TST", -3*3600)
	time.Local = loc

	got, err := parseLocalTimestamp("2026-07-06 09:00")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2026-07-06T09:00:00-03:00" {
		t.Fatalf("got %q", got)
	}
}

func TestParseLocalTimestampRejectsInvalid(t *testing.T) {
	_, err := parseLocalTimestamp("tomorrow morning")
	if err == nil || !strings.Contains(err.Error(), "invalid timestamp") {
		t.Fatalf("expected invalid timestamp error, got %v", err)
	}
}

func TestFormatDurationSeconds(t *testing.T) {
	tests := map[int64]string{
		0:    "0:00:00",
		59:   "0:00:59",
		60:   "0:01:00",
		3661: "1:01:01",
	}
	for input, want := range tests {
		if got := formatDurationSeconds(input); got != want {
			t.Fatalf("formatDurationSeconds(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestFormatMoney(t *testing.T) {
	if got := formatMoney(15000, "BRL"); got != "150.00 BRL" {
		t.Fatalf("got %q", got)
	}
	if got := formatMoney(float64(12345), "USD"); got != "123.45 USD" {
		t.Fatalf("got %q", got)
	}
	if got := formatMoney(nil, "BRL"); got != "" {
		t.Fatalf("got %q", got)
	}
}
