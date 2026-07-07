# Ponto CLI

[🇺🇸 English](README.md) · **🇧🇷 Português**

`ponto` é a interface de linha de comando do [Ponto](https://github.com/alextakitani/ponto),
um time-tracker enxuto e self-hosted (Cliente → Projeto → Tarefa → Lançamento).
Cronometre horas, gerencie seu catálogo e exporte relatórios faturáveis do
terminal ou através de agentes de IA.

- Funciona sozinho ou com qualquer agente de IA (Claude, Codex, Copilot, Gemini)
- Saída em JSON com breadcrumbs pra navegar com facilidade
- Autenticação por token contra a **sua própria instância** (é self-hosted — não
  há servidor padrão)
- Inclui uma skill de agente embutida e setup pro Claude Code
- Binário estático único; fork estrutural do
  [fizzy-cli](https://github.com/basecamp/fizzy-cli) (MIT)

## Início rápido

Até que os binários de release sejam publicados, compile do código (Go 1.26+):

```bash
git clone https://github.com/alextakitani/ponto-cli
cd ponto-cli && make build     # → bin/ponto
ponto setup
```

O assistente de setup pergunta a URL da sua instância (**obrigatória** — ex.:
`https://ponto.exemplo.com`), o seu token de acesso, um perfil nomeado (ex.: `prod`,
`dev`) e um projeto padrão opcional.

Gere um token no app do Ponto em **Preferências → Extensão & CLI**. Tokens são
`read` ou `write`; o valor é mostrado uma única vez. Um token `write` é necessário
pra qualquer coisa além de listar e exportar.

Primeiras verificações recomendadas:

```bash
ponto doctor
ponto timer status
```

Use `ponto doctor` sempre que quiser um diagnóstico completo da instalação,
config, auth, conectividade com a API e setup do agente.

<details>
<summary>Outros métodos de instalação</summary>

**Go install:**
```bash
go install github.com/alextakitani/ponto-cli/cmd/ponto@latest
```

**Script instalador / Homebrew / deb / rpm:** o pipeline do goreleaser e o
`scripts/install.sh` já estão prontos e vão funcionar assim que a primeira release
do GitHub for marcada.

</details>

## Próximos passos

O loop principal — cronometrar o dia todo, faturar no fim do mês:

```bash
ponto timer start --description "Consertando o build"   # usa seu projeto padrão
ponto timer start --project 7 --task 3 --description "Code review"
ponto timer status
ponto timer stop

ponto entry list
ponto entry create --start "2026-07-06 09:00" --end "2026-07-06 10:30" \
  --project 7 --description "Planejamento" --new-tag sprint-42
ponto entry duplicate 42        # reinicia um lançamento finalizado como novo timer
ponto entry split 42 --at "2026-07-06 10:00"

ponto export --period month -o relatorio.csv
ponto export --period custom --from 2026-06-01 --to 2026-06-30 \
  --client 1 --group-by project --format xlsx
```

Gerencie o catálogo:

```bash
ponto client list
ponto client create --name "Acme Corp" --currency BRL --rate-cents 15000
ponto project create --name "Homelab" --client 1 --color "#1e66f5"
ponto project default 7         # timer start sem --project usa este
ponto task create --project 7 --name "Infra"
ponto tag create --name backend
ponto client archive 1          # soft-delete; --archived lista de volta
ponto client unarchive 1
```

Pra ver todos os comandos, rode `ponto commands --json` ou leia
[`skills/ponto/SKILL.md`](skills/ponto/SKILL.md).

### Semânticas do timer que vale conhecer

- Há no máximo **um timer rodando** por usuário — o servidor garante isso.
  Iniciar um segundo devolve um erro claro de "timer já está rodando".
- `timer start` sem `--project` deixa o **servidor** aplicar o seu projeto
  padrão. Use `--no-project` pra iniciar explicitamente sem projeto.
- As rates são congeladas por lançamento no momento em que ele é criado
  (`rate_cents` + `currency`); mudar a rate de um projeto depois nunca reescreve
  o histórico.
- Horários que você digita sem offset são enviados com o offset local da sua
  máquina, então "2026-07-06 09:00" significa o que você espera.

### Formatos de saída

```bash
ponto entry list                                  # Envelope JSON
ponto entry list --jq '.data[0].description'      # jq embutido (sem jq externo)
ponto entry list --quiet                          # Dados crus, sem envelope
ponto entry list --styled                         # Tabelas de terminal pra humanos
ponto entry list --markdown                       # Tabelas Markdown
ponto project list --ids-only                     # Um ID por linha
```

`--jq` implica JSON e não pode ser combinado com `--styled`, `--markdown`,
`--ids-only` ou `--count`.

### Envelope JSON

Todo comando devolve JSON estruturado:

```json
{
  "ok": true,
  "data": [...],
  "summary": "3 projects",
  "breadcrumbs": [{"action": "show", "cmd": "ponto project show <id>"}]
}
```

Os breadcrumbs sugerem os próximos comandos, facilitando a navegação pra humanos
e agentes. A saída de lista/detalhe também carrega campos de apresentação
derivados (`duration` como `H:MM:SS`, `rate` como `"150.00 BRL"`) ao lado dos
valores crus da API (`duration_seconds`, `rate_cents`, `currency`).

## Integração com agentes de IA

`ponto` funciona com qualquer agente de IA que roda comandos de shell — "inicie um
timer no projeto Kube", "quantas horas faturáveis esta semana?", "exporte junho
como xlsx".

**Claude Code:** `ponto setup claude` — linka a skill embutida no diretório de
skills do Claude.

**Outros agentes:** aponte o agente pra
[`skills/ponto/SKILL.md`](skills/ponto/SKILL.md). `ponto skill` abre o instalador
interativo; `ponto skill install` instala direto.

**Descoberta pelo agente:** todo comando aceita `--help --agent` pra ajuda em
formato estruturado. Use `ponto commands --json` pro catálogo completo de comandos.

## Configuração

```
~/.config/ponto/              # Config global
├── config.json               #   Perfis nomeados (URL base, projeto padrão)
├── config.yaml               #   Configurações globais
└── credentials/              #   Storage de token de fallback (sem keyring)

.ponto.yaml                   # Por repositório (config local sobrepõe a global)
```

Prioridade da configuração (da mais alta pra mais baixa):

1. Flags de CLI (`--token`, `--profile`, `--api-url`)
2. Variáveis de ambiente (`PONTO_TOKEN`, `PONTO_PROFILE`, `PONTO_API_URL`)
3. Configurações do perfil nomeado (`config.json`)
4. Config local do projeto (`.ponto.yaml`)
5. Config global (`~/.config/ponto/config.yaml` ou `~/.ponto/config.yaml`)

**Não há `api_url` padrão** — o Ponto é self-hosted, então a URL sempre aponta pra
sua instância. Tokens ficam no keyring do sistema (`PONTO_NO_KEYRING=1` força o
fallback em arquivo); `PONTO_NO_UPDATE_NOTIFIER=1` silencia a checagem de updates.

Perfis tornam várias instâncias indolores — ex.: `prod` (seu homelab) e `dev`
(localhost:3000):

```bash
ponto timer status --profile dev
PONTO_PROFILE=dev ponto entry list
```

Inspecione a config efetiva e a precedência:

```bash
ponto config show
ponto config explain
ponto config explain --profile dev
```

## Solução de problemas

```bash
ponto doctor                 # Diagnóstico completo de install/config/auth/API
ponto doctor --profile dev   # Checa um perfil salvo explicitamente
ponto doctor --all-profiles  # Varre todos os perfis salvos
ponto doctor --verbose       # Inclui detalhes da config efetiva
ponto doctor --json          # Saída estruturada pra scripts
```

Comandos comuns de acompanhamento:

```bash
ponto auth status
ponto config show
ponto config explain
ponto setup
ponto setup claude
ponto skill install
```

Os erros mapeiam pra códigos de saída semânticos (não encontrado, auth, proibido,
rate-limit, rede, API) e incluem um `hint` com o comando que normalmente resolve.

## Desenvolvimento

```bash
make build            # Compila o binário → bin/ponto
make test-unit        # Testes unitários (sem API)
make check            # fmt + vet + lint + tidy + testes com race
make e2e              # Suíte e2e de contrato da CLI contra uma instância real
make surface-check    # Verifica se o SURFACE.txt (snapshot da superfície) está atual
```

Requisitos do e2e (os testes são pulados quando ausentes):

- `PONTO_TEST_TOKEN` — um token `write` numa conta **descartável**
- `PONTO_TEST_API_URL` — a instância contra a qual rodar
- opcional: `PONTO_TEST_BINARY`

Docs pra quem contribui: [`docs/spec.md`](docs/spec.md) (o que esta CLI é e por
quê), [`docs/api.md`](docs/api.md) (o contrato da API JSON do Ponto),
[`docs/fork-plan.md`](docs/fork-plan.md) (o que foi herdado do fizzy-cli).

## Licença

[MIT](LICENSE.md). O app Ponto em si é licenciado sob
[O'Saasy](https://github.com/alextakitani/ponto/blob/main/LICENSE.md); a CLI é MIT
de propósito — a mesma combinação que o fizzy usa.
