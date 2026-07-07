# API JSON do Ponto — Referência para o CLI

> Mapeada em 07/07/2026 explorando `~/Projetos/ponto` (Rails). Fontes
> principais: `app/models/access_token.rb`, `app/controllers/concerns/
> authentication.rb`, `config/routes.rb`, views `*.json.jbuilder`,
> `db/schema.rb`.

## ⚠️ Gaps conhecidos (app-side) — verificado em 07/07/2026

O Bearer token só autentica quando `request.format.json?`
(`authentication.rb:110`). Consequência: **hoje o token NÃO funciona** em:

1. **`GET /reports/export.csv|.xlsx`** — formato csv/xlsx ⇒ bearer nem é
   tentado ⇒ redirect pro login. O comentário em `authentication.rb:27` mostra
   que a intenção existe ("a futura rota de export usam isso"), mas o gate não
   foi estendido. **Pendência no app** antes de `ponto export` existir.
2. **Archival** (`POST|DELETE /{clients,projects,tasks,tags}/:id/archival`) —
   controllers só respondem HTML/Turbo. Q73 pede superfície JSON total;
   **pendência no app** antes de `ponto client archive` etc. existirem.

O CLI v0 fica com o que funciona hoje: timer, time_entries (CRUD + split +
duplicate), clients/projects/tasks/tags (CRUD), default project.

## 1. Autenticação

Bearer token no header, em requests JSON:

```
Authorization: Bearer <token>
Accept: application/json
Content-Type: application/json   # em POST/PATCH com body
```

A auth Bearer só é tentada quando (`authentication.rb:109–111`):
`request.format.json?` **e** o header `Authorization` contém `"Bearer"`.
CSRF é dispensado para Bearer JSON (`request_forgery_protection.rb:22`).

### Modelo `AccessToken` (`app/models/access_token.rb`)

| Campo | Tipo | Descrição |
|---|---|---|
| `token` | string | Opaco (`has_secure_token`), exibido uma única vez na criação |
| `permission` | `"read"` \| `"write"` | GET/HEAD passam com qualquer um; escrita exige `"write"` |
| `label` | string\|null | Nome dado pelo usuário |
| `last_used_at` | datetime\|null | Atualizado a cada request autenticado |

Token `read` tentando escrita → **401** (sem mensagem de escopo). Tokens são
gerados na tela **Preferências → Extensão & CLI** (`/preferences`); o valor em
claro aparece uma única vez.

### Respostas de auth

| Situação | HTTP | Body |
|---|---|---|
| Sem/inválido token | 401 | `{"error": "unauthorized"}` |
| Token `read` em escrita | 401 | `{"error": "unauthorized"}` |
| Usuário suspenso | 403 | `{"error": "account suspended"}` |
| Policy negada | 403 | `{"error": "forbidden"}` |
| Recurso alheio | 404 | (padrão Rails) |

## 2. Convenções gerais

- **Timestamps** de saída: ISO 8601 UTC (`"2026-07-06T14:30:00.000Z"`).
- **Timestamps de entrada** (`started_at`, `ended_at`, `split.cut`): sem
  offset ⇒ interpretados no **fuso do usuário** (Preferências, default
  `America/Sao_Paulo`); com offset/Z ⇒ literal
  (`time_entries_controller.rb:125–128`). **O CLI deve sempre enviar offset
  explícito.**
- **Dinheiro**: sempre `rate_cents` (int) + `currency` (ISO 4217). `null` =
  sem taxa. No body de escrita, enviar `rate_cents` (não `rate`, que é o
  writer do form HTML com parse pt-BR).
- **Erros**: dois formatos — singular `{"error": "..."}` (auth, 409, 404) e
  validação `{"errors": ["...", ...]}` (422). Sem envelope extra.
- **Ordenação**: catálogo por `name_normalized ASC`; entries por
  `started_at DESC`.
- **Paginação**: NÃO há em nenhum endpoint JSON — arrays completos.

## 3. Timer (resource singular)

Invariante: no máximo 1 timer por usuário, reforçada por índice parcial único
(`index_time_entries_running_per_user WHERE ended_at IS NULL`,
`db/schema.rb:181`).

### `GET /timer`

200 com o entry rodando, ou `null` (literal) se não houver.

```json
{
  "id": 42, "project_id": 7, "task_id": null,
  "description": "ingress migration",
  "started_at": "2026-07-06T14:30:00.000Z", "ended_at": null,
  "rate_cents": 15000, "currency": "BRL", "billable": true,
  "duration_seconds": null, "billable_amount_cents": null
}
```

`duration_seconds`/`billable_amount_cents` são sempre `null` com timer rodando.

### `POST /timer`

Body (tudo opcional): `{"timer": {"project_id": 7, "task_id": 3, "description": "..."}}`

- **`project_id` AUSENTE** da chave `timer` (ou chave `timer` omitida) ⇒
  servidor aplica o **projeto padrão ativo** do usuário
  (`timers_controller.rb:33–36`). **`project_id: null` explícito** ⇒ sem
  projeto. O CLI usa isso: sem `--project`, omitir a chave.
- `task_id` deve pertencer ao projeto.
- Snapshot: servidor congela `rate_cents`/`currency` da `effective_rate` do
  projeto; `billable` derivado (`rate_cents != null`). `timer_params` só
  aceita `project_id`, `task_id`, `description`.

| Situação | HTTP | Body |
|---|---|---|
| Sucesso | 201 | entry completo |
| Timer já rodando (inclusive race no índice) | 409 | `{"error": "timer is already running"}` |
| Validação | 422 | `{"errors": [...]}` |

### `DELETE /timer`

| Situação | HTTP | Body |
|---|---|---|
| Parado com duração > 0 | 200 | entry completo finalizado |
| Duração = 0 (entry destruído) | 204 | — |
| Sem timer | 404 | `{"error": "timer not found"}` |

## 4. Time Entries

### Shape (`time_entries/_time_entry.json.jbuilder`)

```jsonc
{
  "id": 43,
  "project_id": 7,            // ou null
  "task_id": null,
  "description": "code review",
  "started_at": "2026-07-06T09:00:00.000Z",
  "ended_at": "2026-07-06T10:30:00.000Z",   // null = rodando
  "rate_cents": 15000,        // snapshot congelado na criação
  "currency": "BRL",          // snapshot
  "billable": true,
  "duration_seconds": 5400,          // null se rodando
  "billable_amount_cents": 22500     // null se rodando, não-billable ou sem rate
}
```

`billable_amount_cents = round_half_up(rate_cents × duration_seconds / 3600)`
(`time_entry.rb:43–47,184–186`). Snapshot NÃO retroage quando a rate do
projeto/cliente muda; é re-congelado se `project_id` mudar num update
(`time_entry.rb:162–172`).

### Endpoints

- **`GET /time_entries`** → 200, array completo (sem filtros/paginação),
  `started_at DESC`. Inclui o timer rodando.
- **`GET /time_entries/:id`** → 200 | 404.
- **`POST /time_entries`** (entry manual) → 201 | 422. Body:

  ```json
  {"time_entry": {
    "project_id": 7, "task_id": 3, "description": "code review",
    "started_at": "2026-07-06T09:00:00-03:00",
    "ended_at": "2026-07-06T10:30:00-03:00",
    "billable": true, "tag_ids": [1, 5], "new_tag_names": ["nova-tag"]
  }}
  ```

  `started_at` obrigatório; `ended_at` omitido cria timer rodando; `billable`
  omitido deriva do projeto; `new_tag_names` cria/acha tags por nome
  normalizado. Validações (422): `ended_at` > `started_at`; sem sobreposição
  com entry finalizado; `task_id` exige `project_id` e mesmo projeto;
  `tag_ids` só do próprio usuário.
- **`PATCH /time_entries/:id`** → 200 | 422 | 404. Mesmos campos; **`ended_at`
  é silenciosamente ignorado em timer rodando** (parar = `DELETE /timer`).
  Mudar `project_id` re-congela a rate.
- **`DELETE /time_entries/:id`** → 204.
- **`POST /time_entries/:id/split`** → 204 | 422. Body
  `{"split": {"cut": "..."}}`; corte estritamente entre started/ended; só em
  entry finalizado. Re-buscar entries depois (não retorna os resultantes).
- **`POST /time_entries/:id/duplicate`** → 201 (novo timer copiando
  project/task/description/billable) | 409 se timer rodando | 422. Body vazio.
  É o "resume" do Toggl.

## 5. Clients

Shape: `{id, name, rate_cents, currency, note, archived_at, created_at, updated_at}`.

- **`GET /clients`** → 200. Params: `archived` (presente ⇒ lista arquivados),
  `q` (busca por nome normalizada).
- **`GET /clients/:id`** → 200 | 404.
- **`POST /clients`** → 201 | 422. Body: `{"client": {"name": "Acme",
  "currency": "BRL", "rate_cents": 15000, "note": "..."}}`. `name` e
  `currency` obrigatórios; nome único por usuário (inclui arquivados,
  normalizado).
- **`PATCH /clients/:id`** → 200 | 422.
- **`DELETE /clients/:id`** → 204 | 422 se tiver projetos
  (`restrict_with_error`).
- **Archival**: HTML-only hoje (ver Gaps).

## 6. Projects

Shape (`projects/_project.json.jbuilder`):

```jsonc
{
  "id": 7, "name": "Kubernetes homelab", "color": "#1e66f5",
  "client_id": 1,
  "rate_cents": null,              // override próprio; null = herda do cliente
  "archived_at": null, "created_at": "...", "updated_at": "...",
  "default": true,                 // projeto padrão do usuário
  "currency": "BRL",               // = effective_currency
  "effective_rate_cents": 15000,   // override do projeto OU rate do cliente
  "effective_currency": "BRL"
}
```

Cascata (`project.rb:61–80`): `effective_rate_cents` = projeto ?? cliente;
`effective_currency` = currency do cliente ?? `"BRL"` (se houver rate) ?? null.

- **`GET /projects`** → 200. Params: `archived`, `q`, `client_id`.
- **`GET /projects/:id`** → 200 com **`tasks` ativas aninhadas** | 404.
- **`POST /projects`** → 201 | 422. Body: `{"project": {"name": "...",
  "client_id": 1, "color": "#1e66f5", "rate_cents": null}}`. `name`
  obrigatório/único; `color` omitida ⇒ servidor escolhe a menos usada da
  paleta de 12.
- **`PATCH /projects/:id`** → 200 | 422.
- **`DELETE /projects/:id`** → 204 (tasks cascateiam).
- **`POST /projects/:id/default`** → 201 `{"default_project_id": 7}`.
- **`DELETE /projects/:id/default`** → 204 (idempotente).
- **Archival**: HTML-only hoje (ver Gaps).

## 7. Tasks (aninhadas em project; shallow p/ show/update/destroy)

Shape: `{id, name, project_id, archived_at, created_at, updated_at}`.

- **`GET /projects/:project_id/tasks`** → 200 (ativas).
- **`GET /tasks/:id`** → 200 | 404.
- **`POST /projects/:project_id/tasks`** → 201 | 422. Body:
  `{"task": {"name": "Infra"}}`. Nome único por PROJETO (inclui arquivadas).
- **`PATCH /tasks/:id`** → 200 | 422.
- **`DELETE /tasks/:id`** → 204.
- **Archival**: HTML-only hoje (ver Gaps).

## 8. Tags

Shape: `{id, name, archived_at, created_at, updated_at}`.

- **`GET /tags`** → 200. Params: `archived`, `q`.
- **`GET /tags/:id`** → 200 | 404.
- **`POST /tags`** → 201 | 422. Body: `{"tag": {"name": "backend"}}`.
- **`PATCH /tags/:id`** → 200 | 422.
- **`DELETE /tags/:id`** → 204 | 422 se em uso (`restrict_with_error`).
- **Archival**: HTML-only hoje (ver Gaps).

## 9. Reports e Export

Report é PORO + tela HTML — **sem endpoint JSON**. Export
(`GET /reports/export.xlsx|.csv`) existe, mas **não é alcançável por Bearer
hoje** (ver Gaps). Params (para quando destravar): `period`
(`today|week|month|year|last_month|last_year|custom`, default `month`),
`from`/`to` (`YYYY-MM-DD`, com `custom`), `client_ids[]`/`project_ids[]`/
`task_ids[]`/`tag_ids[]` (aceitam `"none"`), `billable`, `description`,
`group_by`/`group_by_2` (`project|client|task|tag|description`),
`rounding=on` + `rounding_block` (`5|15|30`) + `rounding_direction`
(`nearest|up|down`), `export_locale` (`pt-BR|en`). Só entries finalizados
entram (`report.rb:52–57`).

## 10. Regras de domínio que o CLI respeita

- **Timer único**: tratar 409 graciosamente (mostrar o timer corrente).
- **Default project**: sem `--project` ⇒ OMITIR a chave `project_id` do POST
  /timer (servidor aplica o default). "Sem projeto" explícito ⇒
  `project_id: null`.
- **Fuso**: sempre enviar timestamps com offset explícito.
- **Sobreposição**: entries finalizados não se sobrepõem ⇒ 422 com o
  intervalo conflitante.
- **Delete vs archive**: Client/Tag têm delete bloqueado se em uso (422);
  Project/Task/Entry deletam sempre. (Archive via CLI espera o gap fechar.)
- **Unicidade de nomes** (via `name_normalized`, sem acento/lowercase):
  clients/projects/tags por usuário; tasks por projeto.

## 11. Mapa de rotas JSON

| Método | Path | JSON? |
|---|---|---|
| GET / POST / DELETE | `/timer` | sim |
| GET / POST | `/time_entries` | sim |
| GET / PATCH / DELETE | `/time_entries/:id` | sim |
| POST | `/time_entries/:id/split` | sim (204) |
| POST | `/time_entries/:id/duplicate` | sim (201) |
| GET / POST | `/clients` · `/projects` · `/tags` | sim |
| GET / PATCH / DELETE | `/clients/:id` · `/projects/:id` · `/tags/:id` | sim |
| POST / DELETE | `/projects/:id/default` | sim |
| GET / POST | `/projects/:project_id/tasks` | sim |
| GET / PATCH / DELETE | `/tasks/:id` | sim |
| POST / DELETE | `/*/:id/archival` | **não** (HTML) |
| GET | `/reports` | **não** (HTML) |
| GET | `/reports/export.{xlsx,csv}` | **não alcançável por Bearer hoje** |

## 12. Exemplos

```bash
# Iniciar timer com projeto padrão do usuário
curl -s -X POST https://ponto.example.com/timer \
  -H "Authorization: Bearer $TOKEN" -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"timer": {"description": "ingress migration"}}'
# → 201 entry | 409 {"error": "timer is already running"}

# Parar
curl -s -X DELETE https://ponto.example.com/timer \
  -H "Authorization: Bearer $TOKEN" -H "Accept: application/json"
# → 200 entry | 204 (duração 0) | 404 {"error": "timer not found"}

# Projetos ativos com busca
curl -s "https://ponto.example.com/projects.json?q=kube" \
  -H "Authorization: Bearer $TOKEN" -H "Accept: application/json"

# Entry manual com tag inline
curl -s -X POST https://ponto.example.com/time_entries \
  -H "Authorization: Bearer $TOKEN" -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"time_entry": {"project_id": 7, "description": "code review",
       "started_at": "2026-07-06T09:00:00-03:00",
       "ended_at": "2026-07-06T10:30:00-03:00",
       "tag_ids": [1], "new_tag_names": ["sprint-42"]}}'
```
