package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The bug these lock down: a stored token was reported as "authenticated"
// without ever asking the server, so a rejected credential looked fine at
// login and at status, and only failed later from whatever tried to use it.

func TestVerifyTokenAcceptedWhenServerServesRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer good-token" {
			t.Errorf("Authorization = %q, want bearer of the token under test", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	check := VerifyToken(context.Background(), srv.URL, "good-token")
	if !check.Accepted() {
		t.Fatalf("Accepted() = false for a 200 response (verdict %v)", check.Verdict)
	}
}

func TestVerifyTokenRejectedOn401And403(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		check := VerifyToken(context.Background(), srv.URL, "bad-token")
		if check.Verdict != TokenRejected {
			t.Errorf("status %d: verdict = %v, want TokenRejected", status, check.Verdict)
		}
		if check.Accepted() {
			t.Errorf("status %d: Accepted() = true for a rejected token", status)
		}
		if check.HTTPStatus != status {
			t.Errorf("status %d: HTTPStatus = %d, want %d", status, check.HTTPStatus, status)
		}
		srv.Close()
	}
}

// An unreachable or broken server says nothing about the credential. Calling
// that "rejected" would send someone off to regenerate a token that was fine.
func TestVerifyTokenUncertainWhenServerUnreachable(t *testing.T) {
	check := VerifyToken(context.Background(), "http://127.0.0.1:59999", "some-token")
	if check.Verdict != TokenUncertain {
		t.Errorf("verdict = %v, want TokenUncertain for an unreachable host", check.Verdict)
	}
	if check.Accepted() {
		t.Error("Accepted() = true for an unreachable server; absence of proof is not proof")
	}
}

func TestVerifyTokenUncertainOnServerError(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
		check := VerifyToken(context.Background(), srv.URL, "some-token")
		if check.Verdict != TokenUncertain {
			t.Errorf("status %d: verdict = %v, want TokenUncertain (an endpoint problem, not a token problem)",
				status, check.Verdict)
		}
		srv.Close()
	}
}

// No token or no URL means no request was made — never claim a verdict.
func TestVerifyTokenUnverifiedWithoutInputs(t *testing.T) {
	cases := []struct{ name, url, token string }{
		{"no token", "https://example.test", ""},
		{"blank token", "https://example.test", "   "},
		{"no url", "", "some-token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := VerifyToken(context.Background(), tc.url, tc.token)
			if check.Verdict != TokenUnverified {
				t.Errorf("verdict = %v, want TokenUnverified", check.Verdict)
			}
			if check.Accepted() {
				t.Error("Accepted() = true without having asked anything")
			}
		})
	}
}
