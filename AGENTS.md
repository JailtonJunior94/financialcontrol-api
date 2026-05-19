<!-- governance-schema: 1.0.0 -->
# Regras para Agentes de IA

Este diretorio centraliza regras para uso com agentes de IA em tarefas reais de analise, alteracao e validacao de codigo.

## Objetivo

Use estas instrucoes para manter consistencia, seguranca e qualidade ao trabalhar com codigo, configuracao, validacao e evolucao de sistemas.

## Arquitetura: monolito

O projeto aparenta ser um monolito unico. A governanca deve privilegiar coesao local, limites de pacote claros e crescimento incremental da estrutura.

Stack detectada: Go.
Frameworks detectados: Fiber.

## Estrutura de Pastas

```
.
.ai_spec_harness.json
.env
.env.example
.github
.github/agents
.github/agents/bugfix.agent.md
.github/agents/prd-writer.agent.md
.github/agents/project-analyzer.agent.md
.github/agents/refactorer.agent.md
.github/agents/reviewer.agent.md
.github/agents/task-executor.agent.md
.github/agents/task-planner.agent.md
.github/agents/technical-specification-writer.agent.md
.github/copilot-instructions.md
.github/skills
.github/skills/agent-governance
.github/skills/agent-governance/SKILL.md
.github/skills/agent-governance/references
.github/skills/agent-governance/references/bug-schema.json
.github/skills/agent-governance/references/ddd.md
.github/skills/agent-governance/references/enforcement-matrix.md
.github/skills/agent-governance/references/error-handling.md
.github/skills/agent-governance/references/messaging.md
.github/skills/agent-governance/references/observability.md
.github/skills/agent-governance/references/persistence.md
.github/skills/agent-governance/references/security-app.md
.github/skills/agent-governance/references/security.md
.github/skills/agent-governance/references/shared-architecture.md
.github/skills/agent-governance/references/shared-lifecycle.md
.github/skills/agent-governance/references/shared-patterns.md
.github/skills/agent-governance/references/shared-testing.md
.github/skills/agent-governance/references/testing.md
.github/skills/agent-governance/scripts
.github/skills/agent-governance/scripts/detect-architecture.sh
.github/skills/agent-governance/scripts/detect-toolchain.sh
.github/skills/agent-governance/triggers
.github/skills/agent-governance/triggers/go.yaml
.github/skills/agent-governance/triggers/node.yaml
.github/skills/agent-governance/triggers/python.yaml
.github/skills/analyze-project
.github/skills/analyze-project/SKILL.md
.github/skills/analyze-project/assets
.github/skills/analyze-project/assets/agents-template.md
.github/skills/analyze-project/assets/ai-tool-template.md
.github/skills/analyze-project/scripts
.github/skills/analyze-project/scripts/generate-governance.sh
.github/skills/bugfix
.github/skills/bugfix/SKILL.md
.github/skills/bugfix/assets
.github/skills/bugfix/assets/bugfix-report-template.md
.github/skills/bugfix/references
.github/skills/bugfix/references/canonical-bug-format.md
.github/skills/bugfix/scripts
.github/skills/bugfix/scripts/validate-bug-input.py
.github/skills/confluence-changelog-publisher
.github/skills/confluence-changelog-publisher/SKILL.md
.github/skills/create-prd
.github/skills/create-prd/SKILL.md
.github/skills/create-prd/assets
.github/skills/create-prd/assets/prd-template.md
.github/skills/create-tasks
.github/skills/create-tasks/SKILL.md
.github/skills/create-tasks/assets
.github/skills/create-tasks/assets/task-template.md
.github/skills/create-tasks/assets/tasks-template.md
.github/skills/create-technical-specification
.github/skills/create-technical-specification/SKILL.md
.github/skills/create-technical-specification/assets
.github/skills/create-technical-specification/assets/adr-template.md
.github/skills/create-technical-specification/assets/techspec-template.md
.github/skills/execute-task
.github/skills/execute-task/SKILL.md
.github/skills/execute-task/assets
.github/skills/execute-task/assets/task-execution-report-template.md
.github/skills/github-diff-changelog-publisher
.github/skills/github-diff-changelog-publisher/SKILL.md
.github/skills/github-pr-comment-triage
.github/skills/github-pr-comment-triage/SKILL.md
.github/skills/github-release-publication-flow
```

## Padrao Arquitetural

Predominio de packages internos coesos, com estrutura orientada por dominio ou componente.

### Fluxo de Dependencias

- Transporte e adapters devem depender de casos de uso ou servicos explicitos, nao do contrario.
- Dominio nao deve conhecer detalhes de HTTP, banco, filas, serializacao ou drivers.
- Infraestrutura pode implementar contratos consumidos pela aplicacao, preservando dependencia para dentro.

### Forma Canonica de Modulo (`internal/modules/<m>/`)

Todo modulo de negocio segue a mesma forma. Divergencias precisam de justificativa explicita.

1. Camadas: `application/` (`dtos`, `usecase`, `usecase/mocks`), `domain/`
   (`entities`, `vos`, `ports`, `ports/mocks`, `errors.go`; `services/` quando
   houver regra de dominio sem estado), `infrastructure/`
   (`http/handlers`, `http/routes`, `persistence/mssql`).
2. `module.go` expoe o quarteto `Deps` / `Module` / `NewModule` / `RegisterHTTP`.
3. `Deps` carrega apenas dependencias externas brutas (`DB devkitdb.DBTX` e/ou
   `Manager manager.Manager`) e adapters cross-module/cross-cutting que **nao
   pertencem** ao modulo. Repositorios, clock e id generator **internos** sao
   construidos dentro de `NewModule` — nunca no container.
4. `infrastructure/http/handlers/error_mapping.go` no pacote `handlers`.
5. `infrastructure/persistence/mssql/`: `queries.go`, `row_mappers.go`,
   `errors.go` (`MapDriverError` traduzindo erro de driver para sentinela de
   dominio quando houver constraint mapeavel) e `<entidade>_repository.go`.
6. Clock como porta de dominio + `infrastructure/clock/system_clock.go` apenas
   quando o modulo precisar de tempo.
7. Naming: `application/dtos/pagination.go`; rotas em `<modulo>_routes.go` (ou
   `<recurso>_routes.go` por recurso).
8. Excecao documentada: `infrastructure/providers/` guarda adapters de portas
   cross-module por design (enforcada em `dependency_rules_test.go` e
   `governance_rules_test.go`). Nao renomear. `infrastructure/adapters/` e para
   adapters cross-cutting (ex.: hasher, token issuer) — proposito distinto.

Divergencias legitimas (sem acao): modulo sem `domain/services`/clock quando nao
ha regra sem estado nem dependencia de tempo; modulo sem paginacao quando nao
lista colecao; subpastas extras de dominio (`filters`, `projections`) conforme
complexidade real.

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

1. Preservar coesao local e dependencia unidirecional entre packages.
2. Evitar helpers transversais que escondam regra de negocio ou IO.
3. Crescer a estrutura apenas quando o codigo atual ja nao comportar a mudanca com clareza.

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

## Denylist de PII (RF-27)

Campos listados abaixo sao automaticamente redacionados em logs, atributos de span e `db.statement` pelo redator em `internal/bootstrap/observability/redactor/denylist.go`.

**Nunca sugira codigo que leia, imprima ou propague esses campos sem passar pelo redator.**

Campos cobertos:

| Categoria | Campos |
|-----------|--------|
| PCI / credenciais | `pan`, `card_number`, `cardnumber`, `cvv`, `cvc`, `password`, `passwd`, `secret`, `token`, `access_token`, `refresh_token`, `authorization` |
| Documentos BR | `cpf`, `cnpj`, `rg` |
| Contato | `email`, `phone`, `telefone`, `address`, `endereco`, `zipcode`, `cep` |
| Bancario e identidade | `account_number`, `agency`, `bank_account`, `pix_key`, `pix_chave`, `full_name`, `holder_name`, `nome_completo` |

Sentinela: `[REDACTED]`. Matching case-insensitive, recursivo em qualquer profundidade, atravessa arrays/listas de objetos.

Evolucao da lista: PR direto + entrada no CHANGELOG. Auditoria trimestral obrigatoria (PR template em `.github/PULL_REQUEST_TEMPLATE/pii-quarterly-review.md`, cron em `.github/workflows/pii-quarterly-review.yml`).

## Validacao

Antes de concluir uma alteracao:

Seguir Etapa 4 de `.agents/skills/agent-governance/SKILL.md` como base canonica.

Comandos detectados no projeto (Go):
1. Rodar fmt: `gofmt -w .`.
2. Rodar test: `go test ./...`.
3. Rodar lint: `golangci-lint run`.

## Restricoes

1. Nao inventar contexto ausente.
2. Nao assumir versao de linguagem, framework ou runtime sem verificar.
3. Nao alterar comportamento publico sem deixar isso explicito.
4. Nao usar exemplos como copia cega; adaptar ao contexto real.


### Controle de profundidade de invocacao

- Skills que invocam outros skills (execute-task, refactor) devem verificar profundidade via `scripts/lib/check-invocation-depth.sh`.
- Limite padrao: 2 niveis. Configuravel via `AI_INVOCATION_MAX`.
- Variaveis de ambiente: `AI_INVOCATION_DEPTH` (corrente), `AI_INVOCATION_MAX` (limite).
