package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestParsePagination(t *testing.T) {
	h := http.Header{}
	h.Set("X-Total-Count", "137")
	h.Set("X-Total-Pages", "3")
	h.Set("X-Page", "2")
	h.Set("X-Per-Page", "50")
	h.Set("X-Next-Page", "3")
	h.Set("X-Prev-Page", "1")

	p := parsePagination(h)
	if p.TotalCount != 137 || p.TotalPages != 3 || p.Page != 2 || p.PerPage != 50 {
		t.Fatalf("unexpected pagination: %+v", p)
	}
	if !p.HasNext() || p.NextPage != 3 || p.PrevPage != 1 {
		t.Fatalf("next/prev wrong: %+v", p)
	}
	if !p.Present() {
		t.Fatalf("Present() should be true")
	}
}

func TestParsePaginationEmpty(t *testing.T) {
	p := parsePagination(http.Header{})
	if p.Present() || p.HasNext() {
		t.Fatalf("empty headers should yield absent pagination: %+v", p)
	}
}

func TestWithPagePreservesParams(t *testing.T) {
	got := withPage("/time_entries?limit=50&since=2026-01-01T00:00:00Z&page=1", 4)
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("withPage produced invalid URL %q: %v", got, err)
	}
	q := u.Query()
	if q.Get("page") != "4" {
		t.Errorf("page not set to 4: %q", got)
	}
	if q.Get("limit") != "50" || q.Get("since") != "2026-01-01T00:00:00Z" {
		t.Errorf("withPage dropped params: %q", got)
	}
	if u.Path != "/time_entries" {
		t.Errorf("withPage mangled path: %q", u.Path)
	}
}

// TestGetWithPaginationFetchAll drives a fake 3-page collection and asserts the
// client walks X-Next-Page, rewriting ?page=, and concatenates every row.
func TestGetWithPaginationFetchAll(t *testing.T) {
	const perPage, total = 2, 5
	var gotPages []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		gotPages = append(gotPages, r.URL.RawQuery)

		totalPages := (total + perPage - 1) / perPage
		start := (page - 1) * perPage
		end := start + perPage
		if end > total {
			end = total
		}

		w.Header().Set("X-Total-Count", strconv.Itoa(total))
		w.Header().Set("X-Total-Pages", strconv.Itoa(totalPages))
		w.Header().Set("X-Page", strconv.Itoa(page))
		w.Header().Set("X-Per-Page", strconv.Itoa(perPage))
		if page < totalPages {
			w.Header().Set("X-Next-Page", strconv.Itoa(page+1))
		}
		if page > 1 {
			w.Header().Set("X-Prev-Page", strconv.Itoa(page-1))
		}

		w.Header().Set("Content-Type", "application/json")
		body := "["
		for i := start; i < end; i++ {
			if i > start {
				body += ","
			}
			body += fmt.Sprintf(`{"id":%d}`, i+1)
		}
		body += "]"
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")

	// Single page: pagination reflects page 1, has a next page.
	one, err := c.GetWithPagination("/time_entries?limit=2", false)
	if err != nil {
		t.Fatalf("single page: %v", err)
	}
	if got := len(one.Data.([]any)); got != perPage {
		t.Fatalf("single page returned %d rows, want %d", got, perPage)
	}
	if !one.Pagination.HasNext() || one.Pagination.NextPage != 2 {
		t.Fatalf("single page pagination wrong: %+v", one.Pagination)
	}

	// Fetch-all: aggregates every row, flattens pagination.
	all, err := c.GetWithPagination("/time_entries?limit=2", true)
	if err != nil {
		t.Fatalf("fetch all: %v", err)
	}
	rows := all.Data.([]any)
	if len(rows) != total {
		t.Fatalf("fetch all returned %d rows, want %d", len(rows), total)
	}
	if all.Pagination.HasNext() {
		t.Fatalf("aggregated pagination should have no next page: %+v", all.Pagination)
	}
	if all.Pagination.TotalCount != total {
		t.Fatalf("aggregated total = %d, want %d", all.Pagination.TotalCount, total)
	}
	// limit param must survive every hop (the fake records RawQuery per request).
	for _, q := range gotPages {
		if v, _ := url.ParseQuery(q); v.Get("limit") != "2" {
			t.Fatalf("a page request dropped limit=2: %q", q)
		}
	}
}
