# Agents guide — ponto-cli

Go CLI for [Ponto](https://github.com/alextakitani/ponto), a self-hosted time
tracker. Structural fork of [fizzy-cli](https://github.com/basecamp/fizzy-cli)
(MIT). Binary name: `ponto`.

## Read first

- `docs/spec.md` — canonical spec (what the CLI is and why).
- `docs/api.md` — the Ponto JSON API this CLI consumes.
- `docs/fork-plan.md` — fizzy-cli map: what is inherited vs replaced.

## Rules

- Code and comments in English. User-facing docs may be in Portuguese.
- Every command outputs the `{ok, data, summary, breadcrumbs}` JSON envelope;
  global flags `--jq`, `--styled`, `--markdown` must keep working.
- Config precedence: flags > env (`PONTO_TOKEN`/`PONTO_API_URL`/`PONTO_PROFILE`)
  > profile > local config > global. `api_url` is REQUIRED (self-hosted; there
  is no default URL).
- Money is scalar: `rate_cents` (int) + `currency` (string).
- Do not work around server invariants (e.g. 409 when a timer is already
  running) — surface them.
- Gates before any commit: `go build ./...`, `go vet ./...`, `go test ./...`.
- Do not touch files outside this repo.
