# ponto-cli — Especificação

> Destilado das decisões de grilling do app Ponto (Q73–Q76, fechadas em
> 02/07/2026 — fonte: `~/Projetos/ponto/docs/grilling-progress.md`). Este
> documento é a spec canônica DO CLI; a referência da API está em
> [api.md](api.md) e o plano de fork em [fork-plan.md](fork-plan.md).

## O que é

CLI em **Go** para o [Ponto](https://github.com/alextakitani/ponto) — time-tracker
self-hosted (Client → Project → Task → TimeEntry). Motivação: "usar o app por
CLI, como o fizzy" — para usar via Claude, ferramentas GUI e clients desktop.

- **Repo separado** (`ponto-cli`), **licença MIT** (o app Ponto é O'Saasy; a
  mesma combinação do fizzy: app O'Saasy + CLI MIT).
- **Fork estrutural do fizzy-cli** (`github.com/basecamp/fizzy-cli`, MIT):
  troca-se domínio e marca, herda-se a infraestrutura pronta.
- Module path: `github.com/alextakitani/ponto-cli`; binário: `ponto`.

## O que se herda do fizzy-cli (Q74)

- Single binary multi-plataforma (goreleaser + installer script).
- Envelope JSON de saída: `{ok, data, summary, breadcrumbs}`.
- `--jq` embutido (gojq), `--styled`, `--markdown`.
- Precedência de config: **flags > env > profile > config local > global**.
- Token em keyring com fallback em arquivo; `setup` interativo; `doctor`.
- Help agent-first: `--help --agent`, `commands --json`.
- Skill/plugin pro Claude.

## Adaptações self-hosted (Q76)

1. **`api_url` OBRIGATÓRIA no setup** — é self-hosted, não existe URL default
   (≠ fizzy.do). Instância do Alex: `https://ponto.takitani.net`.
2. **Profiles** `prod` (homelab) / `dev` (localhost:3000) + env vars
   `PONTO_TOKEN` / `PONTO_API_URL` / `PONTO_PROFILE`.
3. **Skill embarcada + `ponto setup claude`** — o caminho do "usar no Claude".
   Plugin de marketplace fica pra quando for público.
4. **Default project OPCIONAL no perfil** — `timer start` sem `--project` usa o
   default (conveniência, não pré-requisito; espelha a extensão/Q15b).
   Nota: o SERVIDOR também aplica `User#active_default_project` quando o POST
   `/timer` vem **sem** a chave `project_id`; chave presente (mesmo nil/"") é
   respeitada — contrato da API estável para o CLI.

## Auth

O mesmo `AccessToken` da extensão Chrome (bearer, read/write — Q73), gerado na
tela **Preferências → Extensão & CLI** do app. Header `Authorization: Bearer`.

## Superfície de comandos — primeiro corte (Q75)

Incremental: nasce com o que a API já tem (a superfície JSON do app está
pronta — Q73, todo resource de domínio responde JSON nas mesmas rotas via
`respond_to` + jbuilder).

- `timer start / stop / status`
- `entry list / show / create / update / delete`
- Catálogo: `client` / `project` / `task` / `tag` — cada um com
  `list / show / create / update / archive / unarchive`.
- `export` — o app já expõe `GET /reports/export.{csv,xlsx}` via Bearer
  (gap de 07/07 resolvido no mesmo dia — ver [api.md](api.md)).
- Infra (vem do fork em qualquer cenário): `setup`, `doctor`, `commands`,
  skill do Claude.
- `report` (agregados JSON) chega quando o app expuser report em JSON.

**Fora da superfície da API** (Q73): auth de browser (magic-code é fluxo
humano), admin (só-web), import (upload raro, só-web).

## Regras de domínio que o CLI respeita (Q73)

- Dinheiro sempre escalar no JSON: `rate_cents` (int) + `currency` (string),
  nunca objeto Money.
- Invariante de timer único: iniciar com timer rodando → **409**.
- Erros padronizados: `{error: ...}` + status HTTP correto em toda rota JSON.
- Rate é snapshot por lançamento (histórico não revaloriza).
- Corte de dia no fuso do usuário.

## Racional das decisões (por que assim)

- **Go/fork, não Ruby gem ou bash** (Q74): gem/Thor exige runtime e reescreve
  o que já existe pronto; wrapper bash fica aquém do padrão fizzy. Aceito o
  custo da segunda linguagem — padrões prontos, manutenção Alex+Claude.
- **Incremental pós-fatia Timer/TimeEntry** (Q75): o valor pedido é "trackear
  pelo Claude" — dogfood do Ponto via CLI durante o resto do build. Rejeitados:
  big bang pós-export (valor tarde demais) e timer-only (sem catálogo não
  seleciona projeto direito).
