# Changelog

Todas as mudanças notáveis neste projeto são documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/spec/v2.0.0.html) no que se aplica a APIs públicas e contratos operacionais.

## [Unreleased]

### Added

- **PII denylist** auditável em `internal/bootstrap/observability/redactor/denylist.go`. Lista canônica de campos sensíveis cobertos por redator que sanitiza logs, atributos de span e `db.statement`. Evolução via PR direto + entrada nesta seção. Auditoria trimestral obrigatória (PR template + cron Actions).
- Pipeline de observabilidade orientado ao domínio financeiro: counters `financial_operation_total` e `financial_amount_processed`, histograma `partner_integration_latency` (declarado, não alimentado nesta release).
- Driver MSSQL instrumentado em `internal/bootstrap/database/instrumented`: spans OTel por chamada, slow-query log com threshold parametrizável (`SLOW_QUERY_THRESHOLD_MS`, default 1000ms), `application_name=financialcontrol-api-{env}` injetado na connection string.
- Stack local de observabilidade em `deployments/observability/`: `grafana/otel-lgtm:0.7.5` + `otelcol-contrib` com `sqlserverreceiver`; regras PromQL versionadas em `deployments/observability/alerts/`; script idempotente `enable_query_store.sql`.
- Métricas de runtime Go: `go_goroutines`, `go_gc_*`, `process_cpu_seconds_total`, `process_resident_memory_bytes`, `process_open_fds`.
- Resource attributes K8s via downward API: `k8s.namespace`, `k8s.pod.name`, `k8s.container.name`, `service.instance.id` (= `POD_UID`).

### Changed

- **`devkit-go` substituído por fork interno** via `replace` directive em `go.mod` (`github.com/jailtonjunior94/devkit-go v0.4.1-jjr.0`). Motivo: expor `WithExtraLogHandler`/`WithExtraSpanProcessor` para plugar o redator de PII. PR upstream em andamento; `replace` removido quando aceito. Ver ADR-007.
- `service.version` agora é injetado em build-time via `-ldflags "-X main.version=$(git rev-parse --short HEAD)"`. Builds locais sem `-X` reportam `dev`.
- Logs estruturados JSON em **todos** os ambientes (sem texto colorido em dev). Timestamps em UTC.
- Sampling agora respeita envs OTel padrão (`OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG`). Default `parentbased_always_on`.

### Removed

- **Módulo `planning` removido por completo** (`internal/modules/planning/`): comandos CLI de orçamento/saldo/sync (`budget`, `budget-cards-and-others`, `budget-unified`, `budget-full`, `balance`, `budget-category`, `sync`). O módulo estava órfão desde a unificação em `finance` (task 9.0) — não era importado pelo container nem registrado no cobra. Removidos em cascata: `pkg/persistence` (read models consumidos só pelo planning) e o plumbing de CLI associado — interface `CLIRunner` reduzida a `RunServer()`, campo `ModuleRegistration.RegisterCLI` e `modules.RegisterCLI` eliminados.
- **BREAKING:** env `OTEL_EXPORTER_TYPE` (`otlp-grpc`, `otlp-http`, `stdout`) removida. Substituída por `OTEL_EXPORTER_OTLP_PROTOCOL ∈ {grpc, http/protobuf}` (default `grpc`), aderente ao SDK OTel oficial. Manifestos K8s, compose e `.env.example` atualizados no mesmo PR. Ver ADR-005.

### Security

- LGPD/PCI: campos sensíveis (PAN, CVV, senhas, tokens, CPF, CNPJ, RG, e-mail, telefone, endereço, dados bancários, chave PIX, nome completo) são automaticamente redacionados em logs, atributos de span e `db.statement` pela denylist. Sentinela literal: `[REDACTED]`. Guardrail end-to-end em CI (job `pii-guardrail`, bloqueante).
- Auditoria trimestral da denylist: PR de revisão a cada 12 semanas, mesmo sem mudanças (`reviewed, no changes`).

### Deprecated

- (vazio)
