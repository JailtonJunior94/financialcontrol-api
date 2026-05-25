# Financial Control API

[![Go](https://github.com/JailtonJunior94/financialcontrol-api/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/JailtonJunior94/financialcontrol-api/actions/workflows/ci-cd.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=alert_status)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=bugs)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=code_smells)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=coverage)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)

API backend para controle de financas pessoais, organizada em um layout Go-like com entrypoint unico em `cmd/financialcontrol-api`.

## Visao geral da estrutura

```text
.
|-- cmd/
|   |-- financialcontrol-api/   # entrypoint do binario principal
|   `-- test-permissions/       # binario auxiliar de teste
|-- configs/                    # config.<ENV>.yaml
|-- deployments/
|   |-- docker/                 # Dockerfile e compose
|   `-- k8s/                    # manifests Kubernetes
|-- internal/
|   |-- application/            # DTOs, services, handlers e use cases
|   |-- bootstrap/              # wiring de CLI, HTTP e container
|   |-- domain/                 # entidades, eventos e contratos
|   |-- http/                   # middlewares e constantes HTTP compartilhadas
|   |-- infrastructure/         # config, banco, queries e repositories
|   `-- shared/                 # utilitarios internos do servico
|-- tests/                      # artefatos auxiliares de teste e coverage
|-- Makefile
`-- .github/workflows/ci-cd.yml
```

## Entry point e fluxo de execucao

- Binario principal: `cmd/financialcontrol-api/main.go`
- Bootstrap de CLI: `internal/bootstrap/cli`
- Bootstrap HTTP: `internal/bootstrap/http`
- Composicao de dependencias: `internal/bootstrap/container`

O binario sobe a API por padrao e tambem expoe os comandos operacionais `sync`, `budget`, `budget-cards-and-others`, `budget-unified`, `budget-full`, `balance` e `budget-category`.

Para listar os comandos disponiveis:

```bash
go run ./cmd/financialcontrol-api --help
```

## Configuracao

Os arquivos de configuracao ficam em `configs/`:

- `configs/config.development.yaml`
- `configs/config.staging.yaml`
- `configs/config.production.yaml`

O loader resolve o arquivo pelo valor de `ENVIRONMENT` (por ex. `development`) e falha cedo quando o arquivo nao existe.

Exemplos:

```bash
ENVIRONMENT=development go run ./cmd/financialcontrol-api
ENVIRONMENT=production go run ./cmd/financialcontrol-api sync
```

## Variaveis de Ambiente

### Bootstrap HTTP (obrigatorias para subir o servidor)

| Variavel | Formato | Default | Valores aceitos / Notas |
|---|---|---|---|
| `SERVICE_NAME` | string | — (obrigatoria) | Qualquer string nao vazia; ex.: `financialcontrol-api` |
| `SERVICE_VERSION` | string | — (obrigatoria) | Qualquer string nao vazia; ex.: `1.0.0` ou SHA de commit |
| `ENVIRONMENT` | string | — (obrigatoria) | **Exatamente** `development`, `staging` ou `production` (case-sensitive, lowercase) |
| `PORT` | inteiro | `3000` | Porta de escuta do servidor HTTP |
| `HTTP_SHUTDOWN_TIMEOUT` | `time.Duration` | `15s` | Ex.: `15s`, `1m`. Valores nao parseaveis ou `<= 0` abortam o startup |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | string | `grpc` | `grpc` ou `http/protobuf` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | host:port ou URL | `localhost:4317` (grpc) / `localhost:4318` (http) | Endpoint do coletor OTLP |

> **Nota sobre `ENVIRONMENT`:** o novo bootstrap exige o valor em lowercase (`development`, `staging`, `production`). Os arquivos `configs/config.<ENVIRONMENT>.yaml` seguem a mesma convencao em lowercase, resolvendo o startup ponta a ponta com o mesmo valor de env (`config.development.yaml`, `config.staging.yaml`, `config.production.yaml`).

### Configuracao do banco

| Variavel | Formato | Notas |
|---|---|---|
| `MSSQL_CONNECTION_STRING` | string de conexao MSSQL | Ex.: `sqlserver://user:pass@host:1433?database=DB` |

### Pre-requisito: stack LGTM (observabilidade)

O servico **nao sobe sem coletor OTLP acessivel**. Antes de iniciar a API, suba o stack LGTM localmente:

```bash
docker run --rm -p 4317:4317 -p 4318:4318 -p 3000:3000 grafana/otel-lgtm:0.7.5
```

> Atencao: a porta `3000` do container conflita com a porta default da API. Ajuste `PORT` da API para outro valor (ex.: `PORT=8080`) ou altere o mapeamento do container ao executar ambos ao mesmo tempo.

### Endpoints de health check

Apos o servidor subir, os seguintes endpoints sao expostos:

| Endpoint | Descricao |
|---|---|
| `GET /live` | Liveness — sempre retorna 200 |
| `GET /ready` | Readiness — 200 quando MSSQL esta acessivel; 503 caso contrario |
| `GET /health` | Status detalhado em JSON com `service`, `version`, `environment` e estado dos checks |

## Quebras de Contrato

A versao atual introduz duas quebras de contrato visiveis ao consumidor:

### CORS aberto (`*`)

A allowlist anterior (`https://financialcontrol.netlify.app`, `http://localhost:3000`, etc.) foi substituida por `*`. Qualquer origem passa no preflight.

### Payload de erro RFC 7807 (`application/problem+json`)

Respostas de erro agora seguem o formato:

```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "validation failed: name is required",
  "instance": "/api/v1/cards",
  "request_id": "abc-123"
}
```

`Content-Type: application/problem+json` — consumidores que dependiam do formato anterior devem ser atualizados.

## Desenvolvimento local

Build do binario:

```bash
make build
```

Subir a API localmente:

```bash
make run ENVIRONMENT=development
```

Executar sync:

```bash
make run_sync ENVIRONMENT=production
```

Executar comandos de orcamento:

```bash
make run_budget ENVIRONMENT=production DATE=01/04/2025
make run_budget_cards_and_others ENVIRONMENT=production DATE=01/04/2025
make run_budget_unified ENVIRONMENT=production DATE=01/04/2025
make run_budget_full ENVIRONMENT=production DATE=01/04/2025
make run_balance ENVIRONMENT=production DATE=01/04/2025
make run_budget_category ENVIRONMENT=production DATE=01/04/2025 CATEGORY=Alimentacao
```

## Validacao

Rodar a suite de testes unitarios:

```bash
make test
```

Rodar testes de integracao (requer Docker com MSSQL + LGTM acessiveis):

```bash
make test-integration
# equivalente a: go test -tags=integration -timeout=10m ./internal/bootstrap/http/...
```

Rodar validacao estatica disponivel no projeto:

```bash
make vet
```

Gerar HTML de cobertura com base no artefato produzido pelo `make test`:

```bash
go tool cover --html=tests/coverage.out
```

## Docker e operacao

Os artefatos operacionais agora ficam em `deployments/`:

- Dockerfile da API: `deployments/docker/Dockerfile`
- Dockerfile de migracao: `deployments/docker/Dockerfile.migration`
- Compose: `deployments/docker/docker-compose.yml`
- Kubernetes: `deployments/k8s/`

Build da imagem local:

```bash
docker build -f deployments/docker/Dockerfile -t financialcontrol-api:local .
```

Subir stack local com Docker Compose:

```bash
docker compose -f deployments/docker/docker-compose.yml up --build
```

## CI/CD

O workflow em `.github/workflows/ci-cd.yml` executa:

1. `make build`
2. `make test`
3. SonarCloud scan
4. substituicao de variaveis em `configs/config.production.yaml`
5. build e push da imagem a partir de `deployments/docker/Dockerfile`
6. deploy dos manifests em `deployments/k8s/`

## Migrations

O schema do banco e gerenciado pelo binario CLI `cmd/migration`, executado como Init Container ou Job no Kubernetes antes da subida da API.

Consulte o runbook completo em [`cmd/migration/README.md`](cmd/migration/README.md) para:
- Variaveis de ambiente suportadas (`MSSQL_CONNECTION_STRING`, `MIGRATION_TIMEOUT`, `MIGRATION_BASELINE`).
- Exit codes e padroes de log JSON estruturado.
- Procedimento de baseline em ambientes preexistentes.
- Rollback manual via snapshot.
- Recomendacoes de TLS em producao.
- SLA por categoria de migracao e estrategia de rollout faseado.

Build e execucao local:

```bash
make migrate-build          # compila o binario migration
make migrate-up             # aplica migracoes pendentes
make migrate-baseline BASELINE=1  # baseline em ambiente preexistente
```

## Observabilidade

O stack de observabilidade cobre metricas de negocio do dominio financeiro, rastreamento de operacoes MSSQL com atributos OTel, sanitizacao de PII via denylist e alertas PromQL versionados.

### Subir o stack local (Grafana + Tempo + Loki + Mimir + otelcol)

```bash
make observability-up
# equivalente a: docker compose -f deployments/observability/docker-compose.yml up -d
```

Grafana disponivel em `http://localhost:3001` (admin/admin).

Para derrubar o stack:

```bash
make observability-down
```

Para instruir o OTel Collector a coletar DMVs e Query Store do MSSQL, aplique o script idempotente:

```bash
# Via sqlcmd ou ferramenta equivalente apontando ao banco alvo:
sqlcmd -S <host> -U <user> -P <password> -i deployments/observability/mssql/enable_query_store.sql
```

Consulte `deployments/observability/README.md` para detalhes do Collector, mapa de alertas e procedimentos operacionais.

### ADRs desta entrega

| ADR | Decisao |
|-----|---------|
| [ADR-001](tasks/prd-modernization-observability/adr-001-port-financial-metrics-recorder.md) | Port `FinancialMetricsRecorder` no dominio `finance` |
| [ADR-002](tasks/prd-modernization-observability/adr-002-redactor-as-pipeline.md) | Redator de PII como pipeline de atributos |
| [ADR-003](tasks/prd-modernization-observability/adr-003-instrumented-sql-wrapper.md) | Driver wrapper instrumentado para MSSQL |
| [ADR-004](tasks/prd-modernization-observability/adr-004-sqlserverreceiver-collector.md) | `sqlserverreceiver` no OTel Collector para DMVs |
| [ADR-005](tasks/prd-modernization-observability/adr-005-otel-env-evolution.md) | Evolucao das envs OTel (remocao de `OTEL_EXPORTER_TYPE`) |
| [ADR-006](tasks/prd-modernization-observability/adr-006-business-metrics-naming.md) | Nomes de metricas de negocio congelados |
| [ADR-007](tasks/prd-modernization-observability/adr-007-devkit-go-fork.md) | Fork interno do devkit-go via `replace` directive |

## Contrato HTTP

O contrato publico permanece sob o prefixo `/api/v1`. Os testes de bootstrap e de registro no composition point ativo em `internal/bootstrap/http/http_test.go` e `pkg/http/router_test.go` ajudam a garantir a preservacao dos endpoints durante a reorganizacao estrutural.

## Estrategia de Rollout

A migracao do bootstrap HTTP para `devkit-go/server_fiber` e entregue via **deploy direto em producao** — sem feature flag e sem promocao gradual (RF-20).

Checklist pre-merge:
1. `go build ./...` — verde
2. `go test ./...` — verde (testes unitarios)
3. `go test -tags=integration ./...` — verde (testes de integracao, requer Docker)
4. Stack LGTM acessivel antes do deploy (`docker run grafana/otel-lgtm:0.7.5`)
5. Smoke manual: `curl http://localhost:<PORT>/live`, `/ready`, `/health` e uma rota real
6. Postman Collection validada apos importacao

Pos-deploy (T+5 min): validar `/live`, `/ready`, `/health` em producao. Janela de observacao: T+1h monitorando dashboards LGTM e taxa de erros. Rollback: revert do PR e redeploy da versao anterior (sem migracao de dados envolvida).

ADRs desta entrega: [ADR-001](tasks/prd-migration-devkit-httpserver/adr-001-adocao-server-fiber.md), [ADR-002](tasks/prd-migration-devkit-httpserver/adr-002-cors-aberto-rfc7807.md), [ADR-003](tasks/prd-migration-devkit-httpserver/adr-003-fail-fast-observability-mssql.md), [ADR-004](tasks/prd-migration-devkit-httpserver/adr-004-envs-identidade-servico.md).

## AI Governance

Este repositorio usa a governanca `ai-spec` para padronizar o uso de agentes de IA (Claude, Gemini, Codex, Copilot) em tarefas de produto, arquitetura, implementacao e revisao.

- **Ultima atualizacao do baseline:** 2026-05-25
- **Versao `ai-spec`:** 0.23.2 (modo `copy`, linguagem `go`)
- **Ferramentas instaladas:** `claude`, `gemini`, `codex`, `copilot`
- **Fonte canonica de regras:** `AGENTS.md` (raiz). Espelhos por ferramenta: `CLAUDE.md`, `GEMINI.md`. Os fluxos procedurais vivem em `.agents/skills/`.

### Como invocar skills

Skills sao fluxos procedurais executados por um agente de IA. Os mais comuns no ciclo de desenvolvimento:

| Skill | Quando usar |
| --- | --- |
| `create-prd` | Escopar uma nova funcionalidade (objetivo, restricoes, requisitos numerados) |
| `create-technical-specification` | Desenhar arquitetura/interfaces a partir de um PRD aprovado |
| `create-tasks` | Decompor PRD + techspec em tarefas incrementais e testaveis |
| `execute-task` / `execute-all-tasks` | Implementar uma tarefa (ou o PRD inteiro) com validacao e evidencia |
| `review` | Revisar um diff/branch antes do merge |
| `bugfix` | Corrigir bug pela causa raiz com teste de regressao |
| `refactor` | Refatoracao incremental com preservacao de comportamento |

No Claude Code (ou agente compativel), invoque a skill pelo nome (ex.: peca "use a skill `execute-task`"). A skill `go-implementation` e carregada automaticamente para mudancas em codigo Go.

### Verificar a instalacao da governanca

```bash
ai-spec doctor .   # saude geral (git, manifesto, symlinks, permissoes)
ai-spec verify .   # estado das skills (current / missing / drifted)
ai-spec lint .     # validacao dos arquivos de governanca
ai-spec inspect .  # detalhes do manifesto e toolchain detectado
```

> Nota: os hooks `validate-preload` e `validate-governance` precisam ser adicionados manualmente em `.claude/settings.local.json` — o instalador preserva o arquivo existente e nao o sobrescreve.
