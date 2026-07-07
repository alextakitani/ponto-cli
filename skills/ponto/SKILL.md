# Ponto CLI Skill

Use the `ponto` CLI to operate Ponto, a self-hosted time tracker for Clients,
Projects, Tasks, Tags, Timers, and Time Entries.

## Output Contract

Every command returns the standard envelope unless a raw output mode is selected:

```json
{"ok": true, "data": ..., "summary": "...", "breadcrumbs": [...]}
```

Useful global flags:

- `--json`: force the envelope as JSON.
- `--quiet`: print only raw `data`.
- `--jq EXPR`: filter the JSON output with built-in jq syntax.
- `--styled`: render terminal tables.
- `--markdown`: render Markdown tables/details.
- `--limit N`: truncate list output client-side.
- `--profile NAME`, `--api-url URL`, `--token TOKEN`: override configuration.

Configuration precedence is flags, environment, profile, local config, global
config. Environment variables are `PONTO_TOKEN`, `PONTO_API_URL`, and
`PONTO_PROFILE`. `api_url` is required because Ponto is self-hosted.

## Timer

There is at most one running timer per user. The server enforces this; do not
try to work around `409 timer is already running`.

Start with the user's server-side default project:

```bash
ponto timer start --description "code review"
```

Start with an explicit project/task:

```bash
ponto timer start --project 7 --task 3 --description "ingress migration"
```

Start explicitly without a project:

```bash
ponto timer start --no-project
ponto timer start --project 0
```

Important project rule:

- Omit `--project` to omit `project_id` from the request body. The server may
  apply the user's active default project.
- Use `--no-project` or `--project 0` to send `project_id: null`.

Inspect and stop:

```bash
ponto timer status
ponto timer stop
```

`timer status` returns `data: null` and summary `No timer running` when no timer
exists. With a timer, elapsed duration is calculated locally from `started_at`.

`timer stop` can return a finalized entry, or discard a zero-duration timer with
summary `Timer discarded (duration 0:00:00)`. If no timer is running, the CLI
surfaces `no timer running`.

## Entries

List, show, delete:

```bash
ponto entry list
ponto entry show 43
ponto entry delete 43
```

Create a finished manual entry:

```bash
ponto entry create \
  --start "2026-07-06 09:00" \
  --end "2026-07-06 10:30" \
  --project 7 \
  --task 3 \
  --description "code review" \
  --billable \
  --tag 1 \
  --new-tag sprint-42
```

Create a running timer-like entry by omitting `--end`:

```bash
ponto entry create --start "2026-07-06 09:00" --project 7
```

Update only the fields you pass:

```bash
ponto entry update 43 --description "revised note" --not-billable
```

The server ignores `ended_at` when updating a running timer. Use
`ponto timer stop` to stop a timer.

Duplicate an entry into a new timer:

```bash
ponto entry duplicate 43
```

Split a finished entry:

```bash
ponto entry split 43 --at "2026-07-06 09:45"
```

Timestamp rules:

- RFC3339 is accepted: `2026-07-06T09:00:00-03:00`.
- Local common forms are accepted: `2026-07-06 09:00`,
  `2026-07-06 09:00:00`, `2026-07-06T09:00`.
- The CLI always sends timestamps with an explicit offset from the local
  machine timezone.

## Catalog

Clients:

```bash
ponto client list
ponto client list --query acme
ponto client list --archived
ponto client show 1
ponto client create --name Acme --currency BRL --rate-cents 15000 --note "main account"
ponto client update 1 --rate-cents 18000
ponto client delete 1
```

Projects:

```bash
ponto project list
ponto project list --client 1 --query kube
ponto project show 7
ponto project create --name "Kubernetes homelab" --client 1 --color "#1e66f5"
ponto project update 7 --rate-cents 20000
ponto project delete 7
ponto project default 7
ponto project default --clear
```

`project show` includes active tasks nested in the response.

Tasks are nested for list/create and shallow for show/update/delete:

```bash
ponto task list --project 7
ponto task create --project 7 --name Infra
ponto task show 3
ponto task update 3 --name Ops
ponto task delete 3
```

Tags:

```bash
ponto tag list
ponto tag list --query backend
ponto tag list --archived
ponto tag create --name backend
ponto tag update 5 --name infra
ponto tag delete 5
```

Do not invent archive/unarchive commands. The app's archival routes are
HTML-only today and are intentionally not exposed by this CLI.

## Money

Money is scalar in JSON:

```json
{"rate_cents": 15000, "currency": "BRL"}
```

Never send a nested money object or `rate`. Presentation may render this as
`150.00 BRL`, but raw JSON data remains the server's shape.

## Common Agent Patterns

Check configuration and API health:

```bash
ponto doctor --json
ponto config show --json
```

Find IDs:

```bash
ponto client list --jq '.data[] | {id, name}'
ponto project list --query acme --jq '.data[] | {id, name, client_id}'
ponto task list --project 7 --jq '.data[] | {id, name}'
ponto tag list --jq '.data[] | {id, name}'
```

Use server errors as truth. Examples:

- `409 timer is already running`: ask the user whether to inspect or stop it.
- `422`: validation failed; report the server's validation message.
- `401`: token missing/invalid or read token used for write; suggest
  `ponto setup`.
