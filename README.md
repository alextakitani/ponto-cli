# ponto-cli

Command-line interface for [Ponto](https://github.com/alextakitani/ponto) — a
lean, self-hosted time tracker (Client → Project → Task → TimeEntry).

Single Go binary, agent-friendly output, made to be driven by humans and by
AI agents (ships with a Claude skill). Structural fork of
[fizzy-cli](https://github.com/basecamp/fizzy-cli) (MIT).

> **Status: em desenvolvimento** — primeiro corte de comandos (Q75):
> `timer` · `entry` · `client` / `project` / `task` / `tag` (CRUD +
> archive/unarchive) · `export` (CSV/Excel) · `setup` / `doctor` /
> `commands`. `report` (agregados JSON) chega quando o app o expuser.

## Quick start

```sh
ponto setup          # interactive: api_url (required), token, profile
ponto timer start --description "Fixing the build"
ponto timer status
ponto timer stop
ponto entry list --jq '.data[].description'
```

Auth uses the same `AccessToken` as the Chrome extension, generated in the
app under **Preferências → Extensão & CLI** (`Authorization: Bearer`).

Config lives in `~/.config/ponto/` with profiles (e.g. `prod`, `dev`);
precedence is flags > env (`PONTO_TOKEN`, `PONTO_API_URL`, `PONTO_PROFILE`) >
profile > local > global. There is no default `api_url` — it's your instance.

## Output envelope

Every command prints JSON by default:

```json
{ "ok": true, "data": { }, "summary": "Timer started", "breadcrumbs": [] }
```

`--jq EXPR` filters inline (embedded gojq), `--styled` renders for humans,
`--markdown` for docs/agents. `ponto commands --json` and `ponto --help
--agent` expose the full surface to agents.

## Claude

`ponto setup claude` installs the bundled skill so Claude can track time for
you ("start a timer on project X", "how many hours this week?").

## Docs

- [docs/spec.md](docs/spec.md) — what this CLI is and the decisions behind it
- [docs/api.md](docs/api.md) — the Ponto JSON API surface
- [docs/fork-plan.md](docs/fork-plan.md) — fizzy-cli fork map

## License

[MIT](LICENSE.md). The Ponto app itself is O'Saasy; the CLI is MIT on purpose,
mirroring fizzy + fizzy-cli.
