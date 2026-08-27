package commands

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// Verifying a token means asking the server, because only the server knows.
// Everything else — a value in the keyring, a non-empty string in the config —
// says the token was STORED, which is a different claim and the one that used
// to be reported as "authenticated".
//
// `doctor` already asked the right question; `auth status` and `auth login` did
// not, so a rejected token was reported as a working one and the failure only
// surfaced later, at the first real command. This is that single question, in
// one place, for all three callers.

// TokenVerdict is what the server said about a token.
type TokenVerdict int

const (
	// TokenUnverified means no request was made — there was no token or no API
	// URL to send it to. Never claim a token is good or bad on this.
	TokenUnverified TokenVerdict = iota
	TokenAccepted                // the server served an authenticated request
	TokenRejected                // 401/403: the token is not usable
	TokenUncertain               // the server could not be reached, or it errored
)

// TokenCheck carries the verdict plus what is worth telling the user about it.
type TokenCheck struct {
	Verdict    TokenVerdict
	HTTPStatus int
	Latency    time.Duration
	Err        error
}

// Accepted reports whether the server actually served an authenticated request.
// Deliberately false for TokenUncertain: an unreachable server is not evidence
// that a token works.
func (c TokenCheck) Accepted() bool { return c.Verdict == TokenAccepted }

// Reason is a short phrase for humans, matching doctor's existing wording.
func (c TokenCheck) Reason() string {
	switch c.Verdict {
	case TokenAccepted:
		return "Token accepted"
	case TokenRejected:
		return "Token rejected"
	case TokenUncertain:
		if c.Err != nil {
			return "Could not reach the API to verify the token"
		}
		return "The API did not answer the verification request"
	default:
		return "Token not verified"
	}
}

// VerifyToken asks the server whether the token is usable, by making the same
// authenticated request any real command would make.
func VerifyToken(ctx context.Context, apiURL, token string) TokenCheck {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(apiURL) == "" {
		return TokenCheck{Verdict: TokenUnverified}
	}

	base := strings.TrimRight(apiURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/timer", nil)
	if err != nil {
		return TokenCheck{Verdict: TokenUncertain, Err: err}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ponto-cli/"+currentVersion())

	start := time.Now()
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return TokenCheck{Verdict: TokenUncertain, Err: err, Latency: time.Since(start)}
	}
	defer resp.Body.Close()

	check := TokenCheck{HTTPStatus: resp.StatusCode, Latency: time.Since(start)}
	switch {
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		check.Verdict = TokenRejected
	case resp.StatusCode >= 400:
		// A 404 or a 500 says something is wrong with the endpoint, not with the
		// credential. Reporting that as "rejected" would send the user off to
		// regenerate a token that was never the problem.
		check.Verdict = TokenUncertain
	default:
		check.Verdict = TokenAccepted
	}
	return check
}
