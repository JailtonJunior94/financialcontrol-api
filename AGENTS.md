# Regras para Agentes de IA

Este diretorio centraliza regras para uso com agentes de IA em tarefas reais de analise, alteracao e validacao de codigo.

## Objetivo

Use estas instrucoes para manter consistencia, seguranca e qualidade ao trabalhar com codigo, configuracao, validacao e evolucao de sistemas.

## Arquitetura: microservico

O projeto aparenta ser um microservico independente, com foco em contrato de API, inicializacao, dependencias externas e seguranca operacional. A governanca deve preservar o escopo do servico e o seu deploy independente.

Stack detectada: Go.
Frameworks detectados: Fiber.

## Estrutura de Pastas

```
.
cmd
cmd/test-permissions
tasks
tasks/prd-reestruturacao-estrutura-go-like
tasks/prd-reestruturacao-estrutura-go-like/prd.md
tasks/prd-reestruturacao-estrutura-go-like/tasks
tasks/prd-reestruturacao-estrutura-go-like/tasks/02-mover-presentation-e-routes-para-internal-http.md
tasks/prd-reestruturacao-estrutura-go-like/tasks/01-criar-entrypoint-e-bootstrap.md
tasks/prd-reestruturacao-estrutura-go-like/tasks/04-mover-configs-docker-e-manifests.md
tasks/prd-reestruturacao-estrutura-go-like/tasks/03-mover-domain-application-infrastructure-shared.md
tasks/prd-reestruturacao-estrutura-go-like/tasks/05-atualizar-makefile-workflows-docs-e-testes.md
tasks/prd-reestruturacao-estrutura-go-like/adr-001-layout-internal-first.md
tasks/prd-reestruturacao-estrutura-go-like/adr-002-single-binary-bootstrap-split.md
tasks/prd-reestruturacao-estrutura-go-like/techspec.md
tasks/prd-reestruturacao-estrutura-go-like/adr-003-operational-assets-migration.md
go.mod
config.Production.yaml
Dockerfile
Makefile
tests
tests/artillery.yaml
tests/coverage.out
.k8s
.k8s/ingress
.k8s/ingress/ingress.yaml
.k8s/hpas
.k8s/hpas/financialapi-hpa.yaml
.k8s/namespaces
.k8s/namespaces/financialcontrol.yaml
.k8s/certmanager
.k8s/certmanager/issuer.yaml
.k8s/services
.k8s/services/financialapi-svc.yaml
.k8s/deployments
.k8s/deployments/financialapi-dp.yaml
go.sum
docs
readme.md
config.Development.yaml
sonar-project.properties
.gitignore
.env
.github
.github/workflows
.github/workflows/ci-cd.yml
.github/agents
.github/agents/reviewer.agent.md
.github/agents/refactorer.agent.md
.github/agents/task-executor.agent.md
.github/agents/project-analyzer.agent.md
.github/agents/prd-writer.agent.md
.github/agents/technical-specification-writer.agent.md
.github/agents/task-planner.agent.md
.github/copilot-instructions.md
.github/skills
.github/skills/create-prd
.github/skills/analyze-project
.github/skills/execute-task
.github/skills/review
.github/skills/create-tasks
.github/skills/create-technical-specification
.github/skills/refactor
docker-compose.yml
config.Staging.yaml
AGENTS.md
.vscode
.vscode/launch.json
financial_control
main.go
CLAUDE.md
src
src/app
src/app/configuration
src/app/configuration/app.go
src/app/configuration/routes.go
src/app/server.go
src/app/sync.go
src/app/budget.go
src/app/routes
```

## Padrao Arquitetural

Padrao arquitetural nao inferido com alta confianca; assumir composicao simples e dependencias explicitas.

### Fluxo de Dependencias

- Transporte e adapters devem depender de casos de uso ou servicos explicitos, nao do contrario.
- Dominio nao deve conhecer detalhes de HTTP, banco, filas, serializacao ou drivers.
- Infraestrutura pode implementar contratos consumidos pela aplicacao, preservando dependencia para dentro.

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

1. Preservar contratos publicados e compatibilidade de integracao.
2. Manter inicializacao, observabilidade e shutdown como parte do comportamento do servico.
3. Nao acoplar o servico a convencoes de outros servicos sem contrato explicito.

## Regras por Linguagem

Para tarefas que alteram codigo, carregar a skill:

- `.agents/skills/agent-governance/SKILL.md`

Para tarefas que alteram codigo Go, carregar tambem:

- `.agents/skills/go-implementation/SKILL.md`

## Referencias da Skill

Ler conforme necessidade:

- `.agents/skills/agent-governance/references/ddd.md`
- `.agents/skills/agent-governance/references/error-handling.md`
- `.agents/skills/agent-governance/references/security.md`
- `.agents/skills/agent-governance/references/tests.md`

## Referencias da Skill Go

Ler conforme necessidade:

- `.agents/skills/go-implementation/references/governance.md`
- `.agents/skills/go-implementation/references/architecture.md`
- `.agents/skills/go-implementation/references/go-standards.md`
- `.agents/skills/go-implementation/references/interfaces.md`
- `.agents/skills/go-implementation/references/generics.md`
- `.agents/skills/go-implementation/references/concurrency.md`
- `.agents/skills/go-implementation/references/design-patterns.md`
- `.agents/skills/go-implementation/references/observability.md`
- `.agents/skills/go-implementation/references/api.md`
- `.agents/skills/go-implementation/references/persistence.md`
- `.agents/skills/go-implementation/references/configuration.md`
- `.agents/skills/go-implementation/references/implementation-examples.md`

## Validacao

Antes de concluir uma alteracao:

1. Rodar `gofmt` nos arquivos Go alterados.
2. Rodar primeiro testes direcionados e depois `go test ./...` quando o custo for proporcional.
3. Rodar `go vet ./...` quando esse passo fizer parte do gate do projeto.
4. Rodar lint se o contexto oferecer esse passo.
5. Informar falhas com o comando exato e um diagnostico curto.

## Restricoes

1. Nao inventar contexto ausente.
2. Nao assumir versao de linguagem, framework ou runtime sem verificar.
3. Nao alterar comportamento publico sem deixar isso explicito.
4. Nao usar exemplos como copia cega; adaptar ao contexto real.

5. Nao alterar contratos externos, readiness, observabilidade ou semantica operacional sem explicitar a mudanca.
