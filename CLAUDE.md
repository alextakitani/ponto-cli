# ponto-cli

CLI em Go para o Ponto (time-tracker self-hosted, `~/Projetos/ponto`). Fork
estrutural do fizzy-cli (`~/Projetos/fizzy-cli`, MIT). Binário: `ponto`.
Module: `github.com/alextakitani/ponto-cli`. Licença: **MIT** (o app é O'Saasy;
o CLI é MIT de propósito — Q74/Q78).

## Documentação (ler antes de mexer)

- `docs/spec.md` — spec canônica (destilada do grilling Q73–Q76 do app).
- `docs/api.md` — referência da superfície JSON do app Ponto que o CLI consome.
- `docs/fork-plan.md` — mapa do fizzy-cli e o que foi herdado/trocado no fork.

## Convenções

- Código e comentários em **inglês**; docs e conversa em português.
- Saída de comandos segue o envelope `{ok, data, summary, breadcrumbs}` do
  fizzy-cli; `--jq`, `--styled`, `--markdown` disponíveis globalmente.
- Config: flags > env (`PONTO_TOKEN`/`PONTO_API_URL`/`PONTO_PROFILE`) >
  profile > config local > global. `api_url` é obrigatória (self-hosted, sem
  default). Token no keyring com fallback em arquivo.
- Dinheiro é escalar: `rate_cents` int + `currency` string. Nunca parsear/expor
  Money cru.
- Timer: iniciar com timer rodando devolve 409 do servidor — o CLI reporta,
  não contorna. `timer start` sem `--project` omite `project_id` do POST e
  deixa o servidor aplicar o default project do usuário.

## Modo de trabalho

- Fable no main orquestra; implementação → Codex; trivial → direto.
- Gates: `go build ./...`, `go vet ./...`, `go test ./...` verdes antes de
  qualquer commit.
- Instância de referência: `https://ponto.takitani.net` (prod) e
  `localhost:3000` (dev — server do Alex costuma estar de pé).
