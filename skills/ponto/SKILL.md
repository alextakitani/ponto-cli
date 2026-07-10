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

## Pagination

Every list command (`entry list`, `client list`, `project list`, `tag list`,
`task list`) is server-paginated. By default you get **page 1 only** — do not
assume the first response is the whole collection. Three flags control this:

- `--all` — fetch every page and return them merged into one `data` array.
  Use this whenever you need the complete set (counting, summing, finding an
  item by name). The CLI walks the server's next-page links for you.
- `--page N` — fetch one specific page (1-based).
- `--per-page M` — server page size; maps to the API `?limit=`. The server
  caps it at 100, so `--per-page 100` is the largest single request.

```bash
ponto entry list --all                 # every entry, merged
ponto entry list --page 2 --per-page 50
ponto client list --all --jq '.data | length'
```

Do NOT confuse `--per-page` with the global `--limit`: `--limit` only truncates
how many rows are DISPLAYED client-side (and cannot be combined with `--all`);
`--per-page` sets the server page size.

When a paginated response is not the last page, the JSON envelope includes a
`context.pagination` object so you can drive paging programmatically:

```json
{"context": {"pagination": {
  "total": 137, "pages": 3, "page": 1, "per_page": 50,
  "has_next": true, "next": 2, "prev": 0
}}}
```

Detect "more pages exist" with `has_next`; the human summary shows
`N of TOTAL <plural>` when a page is partial. Prefer `--all` over manual paging
unless you specifically need one page.

```bash
# is there more than one page of entries?
ponto entry list --jq '.context.pagination.has_next'
```

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

Filter entries by date window (see also Pagination below). Both bounds are
optional and combinable; `--since` is inclusive and `--until` is EXCLUSIVE
(pass the start of the next period, e.g. next Monday 00:00, to exclude it):

```bash
# entries in July 2026 (upper bound is Aug 1, excluded)
ponto entry list \
  --since "2026-07-01T00:00:00-03:00" \
  --until "2026-08-01T00:00:00-03:00"
```

`--since` / `--until` require a full ISO 8601 timestamp WITH an offset or `Z`.
A bare date like `2026-07-06` is rejected by the server with
`400 invalid since timestamp` — this is intentional (a malformed bound fails
loudly instead of silently returning the wrong window). The filter is applied
before pagination, so the reported total already reflects the window.

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

## Archive / Unarchive

Catalog resources can be archived and restored without deleting historical
references:

```bash
ponto client archive 1
ponto client unarchive 1
ponto project archive 7
```

Tasks and tags use the same verbs:

```bash
ponto task archive 3
ponto task unarchive 3
ponto tag archive 5
```

List archived catalog records when the resource supports an archived list:

```bash
ponto client list --archived
ponto project list --archived
ponto tag list --archived
```

## Export

Export the current month as CSV to the server-provided filename:

```bash
ponto export
```

Export a custom XLSX report to a specific path:

```bash
ponto export --format xlsx --period custom --from 2026-07-01 --to 2026-07-31 --output july.xlsx
```

Stream raw CSV bytes to stdout with no envelope:

```bash
ponto export --period week --client 1 --tag none --output -
```

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
