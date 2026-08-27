# Ponto CLI

**🇺🇸 English** · [🇧🇷 Português](README.pt.md)

`ponto` is a command-line interface for [Ponto](https://github.com/alextakitani/ponto),
a lean, self-hosted time tracker (Client → Project → Task → TimeEntry). Track
time, manage your catalog, and export billable reports from your terminal or
through AI agents.

- Works standalone or with any AI agent (Claude, Codex, Copilot, Gemini)
- JSON output with breadcrumbs for easy navigation
- Token authentication against **your own instance** (it's self-hosted — there
  is no default server)
- Includes an embedded agent skill and Claude Code setup
- Single static binary; structural fork of
  [fizzy-cli](https://github.com/basecamp/fizzy-cli) (MIT)

## Quick Start

Until binary releases are published, build from source (Go 1.26+):

```bash
git clone https://github.com/alextakitani/ponto-cli
cd ponto-cli && make build     # → bin/ponto
ponto setup
```

The setup wizard walks you through your instance URL (**required** — e.g.
`https://ponto.example.com`), your access token, a named profile (e.g. `prod`,
`dev`), and an optional default project.

Get a token in the Ponto app under **Preferências → Extensão & CLI**. Tokens
are `read` or `write`; the value is shown only once. A `write` token is needed
for anything beyond listing and exporting.

Recommended first checks:

```bash
ponto doctor
ponto timer status
```

Use `ponto doctor` any time you want a full health check of your install,
config, auth, API connectivity, and agent setup.

<details>
<summary>Other installation methods</summary>

**Go install:**
```bash
go install github.com/alextakitani/ponto-cli/cmd/ponto@latest
```

**Installer script / Homebrew / deb / rpm:** the goreleaser pipeline and
`scripts/install.sh` are wired up and will work once the first GitHub release
is tagged.

</details>

## Next Steps

The core loop — track time all day, invoice at the end of the month:

```bash
ponto timer start --description "Fixing the build"   # uses your default project
ponto timer start --project 7 --task 3 --description "Code review"
ponto timer status
ponto timer stop

ponto entry list
ponto entry create --start "2026-07-06 09:00" --end "2026-07-06 10:30" \
  --project 7 --description "Planning" --new-tag sprint-42
ponto entry duplicate 42        # restart a finished entry as a new timer
ponto entry split 42 --at "2026-07-06 10:00"

ponto export --period month -o report.csv
ponto export --period custom --from 2026-06-01 --to 2026-06-30 \
  --client 1 --group-by project --format xlsx
```

Manage the catalog:

```bash
ponto client list
ponto client create --name "Acme Corp" --currency BRL --rate-cents 15000
ponto project create --name "Homelab" --client 1 --color "#1e66f5"
ponto project default 7         # timer start without --project uses this
ponto task create --project 7 --name "Infra"
ponto tag create --name backend
ponto client archive 1          # soft-delete; --archived lists them back
ponto client unarchive 1
```

For the full command surface, run `ponto commands --json` or read
[`skills/ponto/SKILL.md`](skills/ponto/SKILL.md).

### Timer semantics worth knowing

- There is at most **one running timer** per user — the server enforces it.
  Starting a second one returns a clear "timer is already running" error.
- `timer start` without `--project` lets the **server** apply your default
  project. Use `--no-project` to explicitly start without one.
- Rates are snapshotted per entry when it's created (`rate_cents` +
  `currency`); changing a project's rate later never rewrites history.
- Timestamps you type without an offset are sent with your machine's local
  offset, so "2026-07-06 09:00" means what you think it means.

### Listing: pagination & date filters

Every list command (`entry`, `client`, `project`, `tag`, `task`) is
**server-paginated** — by default you get the first page, not the whole
collection.

```bash
ponto entry list --all              # fetch every page, merged into one list
ponto entry list --page 2           # one specific page (1-based)
ponto entry list --per-page 100     # server page size (API ?limit=; caps at 100)
```

When a response isn't the last page, the JSON envelope carries a
`context.pagination` object (`total`, `pages`, `page`, `per_page`, `has_next`,
`next`, `prev`) so scripts can page programmatically — but reach for `--all`
unless you specifically need one page. Note `--per-page` (server page size) is
distinct from the global `--limit` (client-side display truncation), and
`--limit` can't be combined with `--all`.

`entry list` also filters by date window, on `started_at`:

```bash
ponto entry list \
  --since "2026-07-01T00:00:00-03:00" \
  --until "2026-08-01T00:00:00-03:00"   # July 2026 (upper bound excluded)
```

`--since` is inclusive, `--until` is **exclusive** (pass the start of the next
period to exclude it). Both need a full ISO 8601 timestamp with an offset or
`Z`; a bare date is rejected by the server with a `400` on purpose.

### Output Formats

```bash
ponto entry list                                  # JSON envelope
ponto entry list --jq '.data[0].description'      # Built-in jq (no external jq needed)
ponto entry list --quiet                          # Raw data, no envelope
ponto entry list --styled                         # Terminal tables for humans
ponto entry list --markdown                       # Markdown tables
ponto project list --ids-only                     # One ID per line
```

`--jq` implies JSON and cannot be combined with `--styled`, `--markdown`,
`--ids-only`, or `--count`.

### JSON Envelope

Every command returns structured JSON:

```json
{
  "ok": true,
  "data": [...],
  "summary": "3 projects",
  "breadcrumbs": [{"action": "show", "cmd": "ponto project show <id>"}]
}
```

Breadcrumbs suggest next commands, making it easy for humans and agents to
navigate. List/detail output also carries derived presentation fields
(`duration` as `H:MM:SS`, `rate` as `"150.00 BRL"`) alongside the raw API
values (`duration_seconds`, `rate_cents`, `currency`).

A failure answers with the same envelope shape, `ok` false, and a machine-
readable `code`:

```json
{
  "ok": false,
  "error": "No API token configured",
  "code": "auth_required",
  "hint": "Run 'ponto auth login TOKEN' or set PONTO_TOKEN"
}
```

Branch on `code`, never on the `error` sentence — the wording is written for
people and can be reworded; the code is the contract.

### Exit codes

Every `code` has a matching exit status, so a shell script can branch without
parsing anything:

| Exit | `code`          | Meaning                                  |
|-----:|-----------------|------------------------------------------|
|  `0` | —               | Success                                  |
|  `1` | `usage`         | Invalid arguments or flags               |
|  `2` | `not_found`     | Resource does not exist                  |
|  `3` | `auth_required` | Not authenticated (missing/invalid token)|
|  `4` | `forbidden`     | Authenticated, but the token lacks scope |
|  `5` | `rate_limit`    | Rate limited (HTTP 429)                  |
|  `6` | `network`       | Connection, DNS or timeout failure       |
|  `7` | `api_error`     | The server returned an error             |
|  `8` | `ambiguous`     | A name matched more than one record      |

Two properties integrators can rely on:

- **A failed command still prints a complete JSON envelope on stdout.** Only a
  missing binary (`127`, from the shell) or a crash produces no JSON. So when
  you parse the output, let the envelope decide and consult the exit status
  only when parsing fails.
- **An unrecognised `code` exits `7`, never `0`.** New codes added later
  degrade to "server returned an error" rather than passing silently.

```bash
ponto --json timer status
case $? in
  0) ;;                                    # running or idle — read .data
  3) echo "run: ponto auth login TOKEN" ;;
  6) echo "instance unreachable" ;;
  *) echo "failed — see .error in the output" ;;
esac
```

### Saved vs. verified credentials

Saving a token and having a working one are different things, and the CLI keeps
them apart:

```bash
ponto auth login TOKEN     # saves, then asks the server whether it works
ponto auth status          # re-checks against the server
```

`token_configured` is a local fact — a token is stored. `authenticated` is a
claim about the server, and is true only after the server has served an
authenticated request. A stored-but-rejected token therefore reports
`authenticated: false` with `verification: "rejected"`, instead of looking fine
here and failing at the first real command.

`auth login` exits `3` when the server rejects the token — the token is still
saved, so a retry with the right value just works. When the instance cannot be
reached, the token is saved and reported as `verification: "unreachable"`: an
unreachable server is not evidence that a credential is bad.

Use `--no-verify` on either command to skip the round trip and report only what
is stored — useful when setting up against an instance you cannot reach yet.
Neither command ever claims `authenticated: true` without having asked.

### Timezones and which day an entry belongs to

The server files an entry under the day its `started_at` falls on **in the
user's timezone**, entirely — an entry that crosses midnight is never split.
JSON timestamps are rendered in the *application's* zone, which is not
necessarily the user's, so a client must not infer the calendar from a
timestamp's own offset, nor from the machine's clock:

```bash
ponto preferences --jq '.data.time_zone'    # e.g. "America/Sao_Paulo"
```

Group by `started_at` converted to that zone and a client's ledger matches the
app's. Use it for the `--since` window too, or a week boundary lands on the
wrong instant for anyone away from the server's zone.

## AI Agent Integration

`ponto` works with any AI agent that can run shell commands — "start a timer
on the Kube project", "how many billable hours this week?", "export June as
xlsx".

**Claude Code:** `ponto setup claude` — links the embedded skill into
Claude's skills directory.

**Other agents:** point your agent at
[`skills/ponto/SKILL.md`](skills/ponto/SKILL.md). `ponto skill` launches the
interactive installer; `ponto skill install` installs directly.

**Agent discovery:** every command supports `--help --agent` for structured
help output. Use `ponto commands --json` for the full command catalog.

## Configuration

```
~/.config/ponto/              # Global config
├── config.json               #   Named profiles (base URL, default project)
├── config.yaml               #   Global settings
└── credentials/              #   Fallback token storage (when keyring unavailable)

.ponto.yaml                   # Per-repo (local config overrides global)
```

Configuration priority (highest to lowest):

1. CLI flags (`--token`, `--profile`, `--api-url`)
2. Environment variables (`PONTO_TOKEN`, `PONTO_PROFILE`, `PONTO_API_URL`)
3. Named profile settings (`config.json`)
4. Local project config (`.ponto.yaml`)
5. Global config (`~/.config/ponto/config.yaml` or `~/.ponto/config.yaml`)

There is **no default `api_url`** — Ponto is self-hosted, so the URL always
points at your instance. Tokens are stored in the system keyring
(`PONTO_NO_KEYRING=1` forces the file fallback);
`PONTO_NO_UPDATE_NOTIFIER=1` silences update checks.

Profiles make multiple instances painless — e.g. `prod` (your homelab) and
`dev` (localhost:3000):

```bash
ponto timer status --profile dev
PONTO_PROFILE=dev ponto entry list
```

Inspect the effective config and precedence:

```bash
ponto config show
ponto config explain
ponto config explain --profile dev
```

## Troubleshooting

```bash
ponto doctor                 # Full install/config/auth/API health check
ponto doctor --profile dev   # Check one saved profile explicitly
ponto doctor --all-profiles  # Sweep every saved profile
ponto doctor --verbose       # Include effective config details
ponto doctor --json          # Structured output for scripts
```

Common follow-up commands:

```bash
ponto auth status
ponto config show
ponto config explain
ponto setup
ponto setup claude
ponto skill install
```

Errors map to semantic [exit codes](#exit-codes) and include a `hint` with the
command that usually fixes it.

## Development

```bash
make build            # Build binary → bin/ponto
make test-unit        # Unit tests (no API required)
make check            # fmt + vet + lint + tidy + race tests
make e2e              # CLI contract e2e suite against a real instance
make surface-check    # Verify SURFACE.txt (CLI surface snapshot) is current
```

E2E requirements (tests skip when unset):

- `PONTO_TEST_TOKEN` — a `write` token on a **disposable** account
- `PONTO_TEST_API_URL` — the instance to run against
- optional: `PONTO_TEST_BINARY`

Docs for contributors: [`docs/spec.md`](docs/spec.md) (what this CLI is and
why), [`docs/api.md`](docs/api.md) (the Ponto JSON API contract),
[`docs/fork-plan.md`](docs/fork-plan.md) (what was inherited from fizzy-cli).

## License

[MIT](LICENSE.md). The Ponto app itself is licensed
[O'Saasy](https://github.com/alextakitani/ponto/blob/main/LICENSE.md); the CLI
is MIT on purpose — the same combination fizzy uses.
