<!-- governance-schema: 1.0.0 -->
# Regras para Agentes de IA

Este diretorio centraliza regras para uso com agentes de IA em tarefas reais de analise, alteracao e validacao de codigo.

## Objetivo

Use estas instrucoes para manter consistencia, seguranca e qualidade ao trabalhar com codigo, configuracao, validacao e evolucao de sistemas.

## Arquitetura: monolito modular

O projeto e um monolito modular: um unico Go module e um unico binario (`cmd/financialcontrol-api`), particionado em bounded contexts autonomos sob `internal/modules/` (`identity`, `cards`, `categories`, `finance`). A governanca deve privilegiar isolamento entre modulos, fronteiras de contexto explicitas e comunicacao cross-module apenas por ports + adapters (anti-corruption layer), nunca por import direto de internals de outro modulo.

Stack detectada: Go 1.26.3.
Frameworks detectados: Fiber v2 (HTTP). Sem gRPC.
Persistencia: SQL Server (go-mssqldb) com migracoes via golang-migrate.
Observabilidade: OpenTelemetry. Auth: JWT. CLI/config: Cobra + Viper.

## Estrutura de Pastas

```
financialcontrol-api/
├── cmd/
│   ├── financialcontrol-api/main.go   # entrypoint da API
│   └── migration/main.go              # entrypoint das migracoes
├── internal/
│   ├── bootstrap/                     # composition root + plataforma
│   │   ├── cli/                       # RunServer (Cobra)
│   │   ├── container/container.go     # monta e injeta os 4 modulos
│   │   ├── config/ · database/ · http/ · health/ · migration/
│   │   └── observability/             # logging, metrics, tracing, redactor
│   └── modules/                       # bounded contexts
│       ├── identity/    (application · domain · infrastructure)
│       ├── cards/       (application · domain · infrastructure)
│       ├── categories/  (application · domain · infrastructure)
│       └── finance/     (application · domain · infrastructure · providers)
│           ├── domain/{entities,vos,services,ports,projections,filters}
│           ├── application/{usecase,dtos}
│           └── infrastructure/{http,persistence/mssql,providers,idempotency,idgen,clock}
├── pkg/                               # shared kernel transversal
│   ├── jwt/ · authmiddleware/ · identityvo/ · identitycontext/
│   ├── customerrors/ · web/ · http/ · events/ · security/ · uuid/
│   └── database/{mssql,migrations} · modules/registration.go
├── configs/ · deployments/ · docs/ · tests/ · scripts/
└── AGENTS.md · CLAUDE.md · GEMINI.md · Makefile   # governanca
```

## Padrao Arquitetural

Clean Architecture / Hexagonal + DDD tatico aplicado por modulo. Cada bounded context expoe a triade `domain/` (entities, vos, services, ports, projections, filters), `application/` (usecase, dtos) e `infrastructure/` (http, persistence/mssql, providers, clock, idgen, idempotency). O wiring de cada modulo segue a forma canonica `NewModule(Deps) *Module` + `RegisterHTTP`, montado no composition root `internal/bootstrap/container/container.go`.

### Fluxo de Dependencias

- Dentro do modulo a dependencia aponta para dentro: `http/handlers -> application/usecase -> domain`. O dominio define interfaces (`domain/ports`); `infrastructure/persistence/mssql` as implementa.
- Dominio nao conhece HTTP, SQL, drivers ou serializacao. Regra verificada por codigo em `internal/modules/finance/domain/dependency_rules_test.go` (proibe `database/sql` e `fiber` no dominio).
- Cross-module so via ports + adapters: o consumidor declara um port no proprio dominio (ex: `finance/domain/ports.CardProvider`) e mapeia para um read model proprio (`finance/domain/projections.CardView`); o adapter vive em `*/infrastructure/providers/` e traduz o contexto de origem. Hoje apenas `finance` consome `cards` e `categories` por esse caminho; `cards`, `categories` e `identity` nao tem dependencia cross-module.
- `pkg/` e o shared kernel transversal (jwt, authmiddleware, customerrors, identityvo, database/mssql, etc.), sem regra de negocio de modulo.

## Modo de trabalho

1. Entender o contexto antes de editar qualquer arquivo.
2. Preferir a menor mudanca segura que resolva a causa raiz.
3. Preservar arquitetura, convencoes e fronteiras ja existentes no contexto analisado.
4. Nao introduzir abstracoes, camadas ou dependencias sem demanda concreta.
5. Atualizar ou adicionar testes quando houver mudanca de comportamento.
6. Rodar validacoes proporcionais a mudanca.
7. Registrar bloqueios e suposicoes explicitamente quando o contexto estiver incompleto.

## Diretrizes de Estrutura

1. Priorize entendimento do codigo e do contexto atual antes de propor refatoracoes.
2. Respeite padroes existentes de nomenclatura, organizacao e tratamento de erro.
3. Defina estrutura simples, evolutiva e com defaults explicitos.
4. Evite reescritas amplas quando uma alteracao localizada resolver o problema.
5. Estabeleca contratos, testes e comandos de validacao cedo quando eles ainda nao existirem.
6. Considere risco de regressao como restricao principal.
7. Evite overengineering disfarcado de arquitetura futura.

## Regras por Arquitetura

Regras de fronteira de modulo (hard):

1. Proibido import direto de internals de outro modulo (`internal/modules/<outro>/...`) fora de `*/infrastructure/providers/`. Comunicacao cross-module so por port (declarado no dominio consumidor) + adapter (na infraestrutura) + projection/read model proprio.
2. Proibido dependencia circular entre bounded contexts. O grafo permitido hoje e `finance -> {cards, categories}`; `identity`, `cards` e `categories` permanecem sem dependencia cross-module.
3. Dominio nao importa HTTP, SQL, drivers nem outro modulo. Replicar `dependency_rules_test.go` ao criar/alterar dominio em `cards`, `categories` e `identity` (hoje so `finance` tem o teste).
4. Wiring novo deve seguir a forma canonica `NewModule(Deps) *Module` + `RegisterHTTP` e ser montado em `internal/bootstrap/container/container.go`, nao em pacotes de dominio/aplicacao.

Regras gerais:

5. Preservar coesao local e dependencia unidirecional entre packages.
6. Evitar helpers transversais que escondam regra de negocio ou IO.
7. Crescer a estrutura apenas quando o codigo atual ja nao comportar a mudanca com clareza.

## Regras por Linguagem

Para tarefas que alteram codigo, carregar a skill:

- `.agents/skills/agent-governance/SKILL.md`

Para tarefas que alteram codigo Go, carregar tambem:

- `.agents/skills/go-implementation/SKILL.md`

Para tarefas de revisao ou refatoracao incremental de design em Go guiadas por heuristicas de object calisthenics, carregar tambem:

- `.agents/skills/object-calisthenics-go/SKILL.md`

Para tarefas de correcao de bugs com remediacao e teste de regressao, carregar tambem:

- `.agents/skills/bugfix/SKILL.md`

### Composicao Multi-Linguagem

Em projetos com mais de uma linguagem (ex: monorepo Go + Node), carregar apenas a skill da linguagem afetada pela mudanca. Se a tarefa cruzar linguagens, carregar ambas e aplicar a validacao de cada stack nos arquivos correspondentes. Nao misturar convencoes de uma linguagem em arquivos de outra.

## Referencias

Cada skill lista suas proprias referencias em `references/` com gatilhos de carregamento no respectivo `SKILL.md`. Nao duplicar a listagem aqui — consultar o SKILL.md da skill ativa para saber quais referencias carregar e em que condicao.

## Notas por Ferramenta

- **Claude Code**: skills pre-carregadas via `.claude/skills/`, hooks via `.claude/hooks/`, agents delegados via `.claude/agents/`.
- **Gemini CLI**: commands em `.gemini/commands/*.toml` apontam para skills canonicas. Sem hooks ou agents nativos — o modelo deve seguir as instrucoes procedurais do SKILL.md carregado.
- **Codex**: le `AGENTS.md` como instrucao de sessao. Entradas em `.codex/config.toml` sao metadados para `upgrade.sh`, nao spec oficial do Codex CLI. O agente deve seguir as instrucoes de `AGENTS.md` para descobrir e carregar skills.
- **Copilot**: `.github/copilot-instructions.md` como instrucao principal. `.github/agents/` sao wrappers. Sem hooks nativos — compliance depende do modelo seguir as instrucoes.

### Matrix de Enforcement

| Capacidade | Claude Code | Gemini CLI | Codex | Copilot |
|---|---|---|---|---|
| Carga base automatica | hook PreToolUse | procedural | procedural | procedural |
| Protecao de governanca | hook PostToolUse | procedural | procedural | procedural |
| Skills pre-carregadas | sim (symlinks) | sim (commands) | nao | sim (agents) |
| Enforcement programatico | sim (hooks) | nao | nao | nao |
| Validacao de evidencias | script | procedural | procedural | procedural |

Ferramentas sem enforcement programatico dependem do modelo seguir instrucoes procedurais. A compliance nessas ferramentas e best-effort.

## Economia de Contexto

Carregar o minimo necessario para a tarefa reduz custo de tokens em 35-50%:

| Complexidade | Criterio | O que carregar |
|---|---|---|
| `trivial` | Rename, typo, import, formatacao | Apenas AGENTS.md |
| `standard` | Bug fix, novo metodo, refactor local | AGENTS.md + TL;DR das references afetadas |
| `complex` | Nova feature, interface publica, migracao | AGENTS.md + referencias completas |

- Classificar a complexidade **antes** de carregar qualquer referencia.
- Quando a reference tiver bloco `<!-- TL;DR ... -->`, preferir o TL;DR ao documento completo em tarefas standard.
- Override explicito via `--complexity=<nivel>` prevalece sobre classificacao automatica.

## Validacao

Antes de concluir uma alteracao:

Seguir Etapa 4 de `.agents/skills/agent-governance/SKILL.md` como base canonica.

Comandos detectados no projeto (Go):
1. Rodar fmt: `gofmt -w .`.
2. Rodar vet: `go vet ./...`.
3. Rodar lint: `golangci-lint run`.
4. Rodar test: `go test ./...`.

## Restricoes

1. Nao inventar contexto ausente.
2. Nao assumir versao de linguagem, framework ou runtime sem verificar.
3. Nao alterar comportamento publico sem deixar isso explicito.
4. Nao usar exemplos como copia cega; adaptar ao contexto real.


### Controle de profundidade de invocacao

- Skills que invocam outros skills (execute-task, refactor) devem verificar profundidade via `scripts/lib/check-invocation-depth.sh`.
- Limite padrao: 2 niveis. Configuravel via `AI_INVOCATION_MAX`.
- Variaveis de ambiente: `AI_INVOCATION_DEPTH` (corrente), `AI_INVOCATION_MAX` (limite).
