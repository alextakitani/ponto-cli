# Plano de fork: fizzy-cli → ponto-cli

> Mapa produzido em 07/07/2026 explorando `~/Projetos/fizzy-cli` (MIT).
> Complementa [spec.md](spec.md). Adaptações-chave pro Ponto, além do renomeio:
>
> - **Sem conceito de account/board**: o Ponto é bolha single-user — caem
>   `FIZZY_ACCOUNT`/`FIZZY_BOARD`, o seletor de conta do setup e o scoping por
>   account slug no client. O "extra" do profile passa a ser o **default
>   project** (opcional — Q76.4).
> - **Sem URL default**: `DefaultAPIURL` some; `api_url` é obrigatória no
>   setup (Q76.1).
> - **Sem SDK gerado**: o `fizzy-sdk/go` não serve; usar o HTTP client legado
>   (`internal/client`) contra a API JSON do Ponto ([api.md](api.md)).
> - **Domínio novo**: timer / entry / client / project / task / tag no lugar
>   de board / card / column / etc.

## 1. Layout do repositório (fizzy-cli)

```
fizzy-cli/
├── cmd/
│   └── fizzy/
│       └── main.go               # Entrypoint: injeta versão via ldflags, chama commands.Execute()
│
├── internal/
│   ├── client/
│   │   ├── client.go             # HTTP client legado (Bearer token, retry 429/5xx, paginação, upload direto)
│   │   ├── interface.go          # Interface API usada para mock em testes
│   │   └── fuzz_test.go          # Fuzzing de parseRetryAfter
│   │
│   ├── commands/                 # Pacote central: toda a lógica de comandos Cobra
│   │   ├── root.go               # Comando raiz, flags globais, PersistentPreRunE, resolução de config/credstore/perfil/SDK
│   │   ├── commands.go           # `fizzy commands [filter]`, catálogo de comandos, agentHelp()
│   │   ├── help.go               # Renderização de help humano (grupos, exemplos, SEE ALSO)
│   │   ├── jq.go                 # jqWriter: filtro gojq inline aplicado sobre o JSON de saída
│   │   ├── markdown.go           # render helpers de markdown para saídas --markdown
│   │   ├── board.go              # DOMÍNIO: board list/show/create/update/delete/publish/... (DESCARTAR)
│   │   ├── card.go               # DOMÍNIO: card CRUD + status/actions (DESCARTAR)
│   │   ├── column.go             # DOMÍNIO: column CRUD/move (DESCARTAR)
│   │   ├── comment.go            # DOMÍNIO: comment CRUD (DESCARTAR)
│   │   ├── step.go               # DOMÍNIO: step CRUD (DESCARTAR)
│   │   ├── reaction.go           # DOMÍNIO: reactions (DESCARTAR)
│   │   ├── tag.go                # DOMÍNIO: tag list (fizzy) — reescrever pro domínio ponto
│   │   ├── webhook.go            # DOMÍNIO: webhooks (DESCARTAR)
│   │   ├── pin.go                # DOMÍNIO: pins (DESCARTAR)
│   │   ├── activity.go           # DOMÍNIO: activity feed (DESCARTAR)
│   │   ├── search.go             # DOMÍNIO: FTS sobre cards (DESCARTAR)
│   │   ├── migrate.go            # DOMÍNIO: migração de board (DESCARTAR)
│   │   ├── attachment.go         # DOMÍNIO: attachments (DESCARTAR)
│   │   ├── inline_attachments.go # DOMÍNIO: --attach helper (DESCARTAR)
│   │   ├── upload.go             # Active Storage upload (DESCARTAR)
│   │   ├── account.go            # account show/settings/export/join-code (DESCARTAR)
│   │   ├── identity.go           # identity show/timezone-update (DESCARTAR)
│   │   ├── user.go               # user management fizzy (DESCARTAR)
│   │   ├── notification.go       # notificações fizzy (DESCARTAR)
│   │   ├── token.go              # token sub-command (DESCARTAR)
│   │   ├── auth.go               # auth login/logout/status/list/switch (MANTER)
│   │   ├── setup.go              # Wizard interativo (huh): token→account→board→save (ADAPTAR)
│   │   ├── setup_agents.go       # setup claude (marketplace+plugin+skill symlink) (MANTER/ADAPTAR)
│   │   ├── signup.go             # signup magic-link fizzy (DESCARTAR)
│   │   ├── skill.go              # skill print/install (embed.FS → ~/.agents/skills/) (MANTER/ADAPTAR)
│   │   ├── doctor.go             # ~20 checks de saúde (MANTER/ADAPTAR checks)
│   │   ├── config_cmd.go         # config show/explain (MANTER)
│   │   ├── update.go             # Background update check (MANTER, repo novo)
│   │   ├── completion.go         # Shell completion (MANTER)
│   │   ├── version.go            # version (MANTER)
│   │   ├── sdk_errors.go         # Erros → output.Error com exit codes (ADAPTAR)
│   │   ├── pseudocolumns.go      # Pseudo-colunas kanban (DESCARTAR)
│   │   ├── columns.go            # Definições de colunas de tabela p/ StyledList (ADAPTAR)
│   │   └── column_color.go       # Validação de cor HEX (DESCARTAR)
│   │
│   ├── config/
│   │   └── config.go             # Load/Save YAML, env, config local vs global, StateDir (MANTER)
│   ├── errors/
│   │   └── errors.go             # Tipos de erro com exit codes (MANTER)
│   ├── harness/
│   │   ├── agent.go              # Registry de agentes de IA (MANTER)
│   │   └── claude.go             # DetectClaude, CheckClaudePlugin/SkillLink (MANTER/ADAPTAR)
│   ├── render/
│   │   ├── render.go             # StyledList/StyledDetail/StyledSummary — lipgloss (MANTER)
│   │   └── markdown.go           # MarkdownList/MarkdownDetail (MANTER)
│   └── tui/
│       ├── banner.go             # ASCII art "FIZZY" (SUBSTITUIR por "PONTO" ou remover)
│       └── banner_anim.go        # Animação do banner no setup (idem)
│
├── e2e/
│   ├── cli_tests/                # E2E executando o binário real contra API real (REESCREVER p/ domínio)
│   ├── harness/                  # Harness genérico: executa binário, parseia envelope (MANTER)
│   └── testdata/fixtures/
│
├── skills/
│   ├── embed.go                  # `//go:embed fizzy` → skills.FS (ADAPTAR)
│   └── fizzy/SKILL.md            # Skill do agente, 1.158 linhas, embarcada (REESCREVER p/ ponto)
│
├── .claude-plugin/               # plugin.json + hooks (ADAPTAR nome/urls)
├── scripts/
│   ├── install.sh                # Installer curl-pipe com SHA256 (ADAPTAR repo)
│   ├── release.sh                # Tag + goreleaser
│   └── publish-aur.sh            # AUR (opcional manter)
├── Makefile                      # build, test-unit, e2e, lint, check, release, surface-* (MANTER)
├── .goreleaser.yaml              # Multi-plataforma, brew/scoop/deb/rpm, cosign (ADAPTAR)
├── go.mod                        # module github.com/basecamp/fizzy-cli
└── SURFACE.txt                   # Snapshot do contrato de superfície CLI (regenerar)
```

## 2. Framework CLI

**Lib:** `spf13/cobra v1.10.2` + `spf13/pflag`. Cada arquivo de domínio tem
`init()` que faz `rootCmd.AddCommand(...)` e registra flags locais. Raiz em
`internal/commands/root.go:73`.

Padrão de um comando de listagem (`board.go:42`):

```go
var boardListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all boards",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := requireAuthAndAccount(); err != nil { return err }
        data, resp, err := getSDK().Boards().List(cmd.Context(), path)
        if err != nil { return convertSDKError(err) }
        items = normalizeAny(data)
        summary := fmt.Sprintf("%d boards", count)
        breadcrumbs := []Breadcrumb{...}
        printListPaginated(items, boardColumns, hasNext, linkNext, all, summary, breadcrumbs)
        return nil
    },
}
```

Mutações: SDK/client tipado → `normalizeAny(data)` → `printMutation(...)` ou
`printMutationWithLocation(...)`.

**PersistentPreRunE** (`root.go:79`), antes de todo comando: valida `--jq`,
resolve formato de saída, carrega config, inicializa credstore e profile
store, `resolveProfile()`, `resolveToken()`, inicializa o client, dispara
update check em background.

## 3. Envelope JSON `{ok, data, summary, breadcrumbs}`

Implementado na lib externa **`github.com/basecamp/cli`** (pacote `output`) —
o fizzy-cli só usa `output.Writer`. Estrutura (cf. `e2e/harness/harness.go:40`):

```json
{
  "ok": true,
  "data": { },
  "summary": "5 boards",
  "breadcrumbs": [{"action":"show","cmd":"fizzy board show <id>","description":"View board details"}],
  "context": {"pagination": {"has_next": true, "next_url": "..."}},
  "meta": {}
}
```

Helpers em `root.go:661–673`: `breadcrumb(action, cmd, desc)`,
`printSuccessWithBreadcrumbs(data, summary, breadcrumbs)` →
`out.OK(data, output.WithBreadcrumbs(...), output.WithSummary(...))`.
`printList` / `printListPaginated` / `printDetail` / `printMutation` fazem
dispatch por formato (`out.EffectiveFormat()`).

- **`--jq`** (`jq.go`): `jqWriter` intercepta o `io.Writer`; usa
  `itchyny/gojq` (sem jq externo). Compilado no PersistentPreRunE. Envelopes
  de erro (`ok:false`) passam sem filtro.
- **`--styled`**: `render.StyledList/StyledDetail` (lipgloss tables).
- **`--markdown`**: `render.MarkdownList/MarkdownDetail` (goldmark).
- **Exclusões** (`root.go:248`, `resolveFormat()`): `--json`/`--quiet`/
  `--ids-only`/`--count`/`--styled`/`--markdown` mutuamente exclusivos;
  `--jq` incompatível com styled/markdown/ids-only/count.

## 4. Config: precedência e armazenamento

Precedência (`config/config.go:112`, `root.go:1148`):

| Camada | Mecanismo |
|--------|-----------|
| 1. Flags | `--token`, `--profile`, `--api-url` |
| 2. Env | `FIZZY_TOKEN`, `FIZZY_PROFILE`, `FIZZY_API_URL`, ... |
| 3. Profile | `~/.config/fizzy/config.json` via `basecamp/cli/profile` |
| 4. Config local | `.fizzy.yaml` (busca do cwd para cima) |
| 5. Config global | `~/.config/fizzy/config.yaml` ou `~/.fizzy/config.yaml` |
| 6. Defaults | `https://app.fizzy.do` ← **no ponto-cli NÃO EXISTE default** |

- **Keyring**: `zalando/go-keyring` via `basecamp/cli/credstore`; chave
  `profile:<name>`; fallback em `~/.config/fizzy/credentials/`; opt-out
  `FIZZY_NO_KEYRING=1`.
- **Setup** (`setup.go`): forms TUI com `charmbracelet/huh`. Fluxo fizzy:
  token (validado contra API) → conta → board → onde salvar → keyring →
  profile → `setupAgents()`. No ponto: **api_url primeiro (obrigatória)** →
  token → default project (opcional) → salvar.
- **Doctor** (`doctor.go`): ~20 checks `{name, status: pass|warn|fail|skip,
  message, hint}` — versão, configs, profile store, filesystem, completion,
  skill, agentes detectados, credenciais, API URL, conectividade, auth.

## 5. HTTP client

Dois coexistem no fizzy-cli:

- **Client legado** (`internal/client/client.go`) — o que o ponto-cli vai
  usar: `Client{BaseURL, Token, HTTPClient, Verbose, Sleeper}`, header
  `Authorization: Bearer <token>`, timeout 30s.
  **Retry** (`client.go:257`): máx 3 tentativas; 429 respeita `Retry-After`
  (cap 300s); 5xx backoff exponencial (1s/2s/4s) só em métodos idempotentes;
  POST/PATCH só retenta em 429.
- **SDK gerado** (`fizzy-sdk/go`) — específico do fizzy, **sai do fork**.

**Erros** (`sdk_errors.go`): mapeia para `output.Error` com exit codes
semânticos (2=not found, 3=auth, 4=forbidden, 5=rate-limit, 6=network,
7=API), com hints acionáveis (`fizzy auth login`, `fizzy doctor`).

## 6. Help agent-first

- `--agent` (`root.go:388`): global; ativa `FormatQuiet` e desativa prompts.
- `--help --agent` (`commands.go:274`): `installAgentHelp()` sobrescreve o
  `SetHelpFunc` do cobra → `agentHelp()` emite JSON `{name, description,
  flags[], subcommands[]}`.
- `commands --json` (`commands.go:65`): `walkCommands()` percorre a árvore e
  emite o catálogo completo (flags via `collectFlags()`).
- `SURFACE.txt`: snapshot por `make surface-snapshot`
  (`TestGenerateSurfaceSnapshot`); CI valida drift com `make surface-check`.

## 7. Skill / plugin Claude

- `skills/embed.go`: `//go:embed fizzy` → `skills.FS`; `SKILL.md` embarcado
  no binário.
- `skill install` (`skill.go`): instala em `~/.agents/skills/<nome>/SKILL.md`
  + symlink `~/.claude/skills/<nome>/`; fallback copia. Locais previstos
  também para opencode e codex (`skill.go:27`).
- **Auto-refresh** (`skill.go:356`): em `PersistentPostRunE`, compara versão
  com `~/.config/<nome>/.last-run-version` e reescreve skills instaladas.
- `setup claude` (`setup_agents.go:42`): `claude plugin marketplace add` +
  `claude plugin install` via exec; repara symlink. No ponto-cli, a Q76.3
  manda: **skill embarcada + `ponto setup claude`**; plugin de marketplace
  fica pra quando for público.
- Detecção (`harness/claude.go`): lê `~/.claude/plugins/installed_plugins.json`.

## 8. Build / release

| Target | O que faz |
|--------|-----------|
| `make build` | `go build -ldflags "-X main.version=<git-tag>"` → `bin/fizzy` |
| `make test-unit` | `go test -v ./internal/...` (sem API) |
| `make e2e` | build + E2E contra API real (`FIZZY_TEST_TOKEN`, `FIZZY_TEST_ACCOUNT`) |
| `make check` | fmt-check + vet + lint + tidy-check + race-test (gate de CI) |
| `make release-check` | check + replace-check + vuln |
| `make surface-snapshot` / `surface-check` | SURFACE.txt |

Goreleaser: CGO_ENABLED=0, linux/darwin/windows/freebsd/openbsd ×
amd64/arm64; brew cask, scoop, deb/rpm (nfpm), checksums + cosign, notarize
macOS. Installer `scripts/install.sh` (curl-pipe, SHA256, default
`~/.local/bin`). E2E: `TestMain` cria `SharedFixture` contra servidor real e
isola config num `HOME` temporário.

## 9. Inventário de renomeio (checklist)

### 9.1 Module path

- `go.mod:1` `module github.com/basecamp/fizzy-cli` → `github.com/alextakitani/ponto-cli`
- `go.mod` manter `github.com/basecamp/cli` (infra genérica: output/profile/credstore)
- `go.mod` **remover** `github.com/basecamp/fizzy-sdk/go`

### 9.2 Binário e nome

- `cmd/fizzy/` → `cmd/ponto/`; `Makefile` `BINARY := bin/ponto`
- `.goreleaser.yaml` project_name/id/main/binary → `ponto`
- `root.go:74` `Use: "ponto"`; strings "Fizzy CLI" → "Ponto CLI"; `version.go`

### 9.3 Env vars (config.go, root.go, doctor.go, config_cmd.go, sdk_errors.go, update.go)

`FIZZY_TOKEN→PONTO_TOKEN`, `FIZZY_PROFILE→PONTO_PROFILE`,
`FIZZY_API_URL→PONTO_API_URL`, `FIZZY_DEBUG→PONTO_DEBUG`,
`FIZZY_NO_KEYRING→PONTO_NO_KEYRING`,
`FIZZY_NO_UPDATE_NOTIFIER→PONTO_NO_UPDATE_NOTIFIER`,
`FIZZY_TEST_*→PONTO_TEST_*` (e2e). **Caem**: `FIZZY_ACCOUNT`, `FIZZY_BOARD`
(sem equivalente — bolha single-user; o extra do profile é o default project),
`FIZZY_ANIM`/`FIZZY_BANNER` (se o banner sair).

### 9.4 Paths de config

- `config.go:16` `DefaultAPIURL` → **remover** (api_url obrigatória)
- `config.go:19` `.fizzy.yaml` → `.ponto.yaml`
- `config.go:66-68` `~/.config/fizzy/` → `~/.config/ponto/`; `~/.fizzy/` → `~/.ponto/`
- `config.go:257` StateDir `~/.local/state/fizzy` → `ponto`
- `root.go:134-145` keyring `ServiceName: "ponto"`, fallback dir, profile store path
- `skill.go:27-34` paths de skill `fizzy` → `ponto`
- `harness/claude.go:30,33,124` marketplace/plugin/skill → ponto

### 9.5 URLs e repo

- `update.go:24` `basecamp/fizzy-cli` → `alextakitani/ponto-cli`
- `doctor.go:324-325` URLs de release
- `.goreleaser.yaml`, `scripts/install.sh:4`, `.claude-plugin/plugin.json`

### 9.6–9.8 Outros

- Keyring service name (`root.go:136`) → `"ponto"`.
- Banner TUI (`tui/banner.go`): ASCII art é o logo FIZZY — substituir ou remover.
- SDK: substituir todos os imports do `fizzy-sdk` pelo client legado apontando
  pra API do Ponto.

## 10. Domínio: descartar vs. manter

**Descartar** (domínio kanban): board, card, column(+colors,+pseudocolumns),
comment, step, reaction, pin, webhook, activity, search, migrate, attachment,
inline_attachments, upload, account, identity, user, notification, token,
signup, banner TUI, SKILL.md do fizzy, SDK, testes E2E de domínio.

**Manter** (infra): root.go (framework todo), commands.go/help.go/jq.go/
markdown.go, setup.go (adaptado), setup_agents.go, skill.go, doctor.go
(checks adaptados), auth.go, config_cmd.go, update.go, completion.go,
version.go, sdk_errors.go (adaptado), config/, errors/, harness/, render/,
client/client.go, e2e/harness/, Makefile, .goreleaser.yaml, install.sh,
.claude-plugin/, mecanismo SURFACE.txt.

**Criar** (domínio ponto, seguindo o padrão da seção 2): `timer start/stop/
status` · `entry list/show/create/update/delete` · `client`/`project`/`task`/
`tag` com `list/show/create/update/archive/unarchive` · SKILL.md novo ·
testes E2E novos.
