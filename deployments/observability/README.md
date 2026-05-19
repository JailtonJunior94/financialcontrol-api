# Observability Stack — `deployments/observability/`

Local observability stack for `financialcontrol-api`.
Provides Grafana, Loki, Tempo and Prometheus (Mimir) via `grafana/otel-lgtm:0.7.5`,
plus a dedicated `otelcol-contrib` sidecar for MSSQL metrics.

---

## Quick Start

```bash
# From the repository root:
make observability-up

# Or directly:
docker compose -f deployments/observability/docker-compose.yml up -d
```

### Endpoints

| Service | URL |
|---------|-----|
| Grafana UI | http://localhost:3001 (admin / admin) |
| OTLP gRPC  | localhost:4317 |
| OTLP HTTP  | localhost:4318 |

Configure the application to export telemetry:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

### Stop / Clean Up

```bash
docker compose -f deployments/observability/docker-compose.yml down
# Remove volumes too (clears Grafana state):
docker compose -f deployments/observability/docker-compose.yml down -v
```

---

## Architecture

```
financialcontrol-api
    │ OTLP gRPC (traces + metrics + logs)
    ▼
grafana/otel-lgtm:0.7.5  ◄── otelcol-contrib
    ├── Grafana  :3000        │ sqlserverreceiver
    ├── Tempo    (traces)     │   ↓ DMVs + Query Store
    ├── Loki     (logs)       │ MSSQL :1433
    └── Mimir    (metrics)
```

> **R2 Verification:** `grafana/otel-lgtm:0.7.5` bundles `otelcol-core`, which does **not**
> include `sqlserverreceiver`. A separate `otelcol-contrib` container is therefore required
> (already wired in `docker-compose.yml`). Confirmed: `otelcol-contrib:0.103.0` ships with
> `sqlserverreceiver` — visible in collector startup logs under `"receiver_type": "sqlserver"`.

---

## MSSQL Setup

### 1. Create the read-only metrics user

Run once against the `financial_control` database:

```bash
# Using sqlcmd from within the mssql container (adjust credentials):
docker exec -it mssql /opt/mssql-tools18/bin/sqlcmd \
  -No -S localhost -U sa -P '@docker@2021' \
  -d financial_control \
  -v OBS_READER_PASSWORD='<strong-password>' \
  -i deployments/observability/mssql/grant_obs_reader.sql
```

Then set the collector environment variables:

```bash
export MSSQL_METRICS_USER=obs_reader
export MSSQL_METRICS_PASSWORD=<strong-password>
```

### 2. Enable Query Store (idempotent)

```bash
docker exec -it mssql /opt/mssql-tools18/bin/sqlcmd \
  -No -S localhost -U sa -P '@docker@2021' \
  -d financial_control \
  -i deployments/observability/mssql/enable_query_store.sql
```

- Requires SQL Server 2017+ (ProductMajorVersion ≥ 14).
- **Idempotent:** second execution prints `"Query Store already in READ_WRITE state"` and exits without error.
- To verify manually: `SELECT actual_state_desc FROM sys.database_query_store_options;`

---

## Alert Map

| File | Alert | Condition | Severity |
|------|-------|-----------|----------|
| `alerts/http.yaml` | `HighTechnicalErrorRate` | 5xx > 1% of requests over 5 min, by route | critical |
| `alerts/http.yaml` | `PaymentRouteHighLatencyP99` | P99 latency > 2s on `POST/PUT/PATCH/DELETE /api/v1/finance/*` over 5 min | critical |
| `alerts/database.yaml` | `MSSQLDeadlocksRecurring` | Deadlocks accumulating over 10 min, sustained ≥ 5 min | critical |
| `alerts/business.yaml` | *(placeholder)* | Business-metric alerts reserved for a future PRD | — |

All groups use `evaluation_interval: 30s`, `for: 5m`, `runbook_url: "TBD"`.

### Validate alert rules locally

```bash
# Using promtool (install via: go install github.com/prometheus/prometheus/cmd/promtool@latest)
promtool check rules deployments/observability/alerts/*.yaml

# Or via Docker:
docker run --rm \
  -v "$(pwd)/deployments/observability/alerts:/alerts:ro" \
  prom/prometheus:v2.52.0 promtool check rules \
  /alerts/http.yaml /alerts/database.yaml /alerts/business.yaml
```

---

## Loading Alerts into Grafana / Mimir

The alert YAML files follow the Prometheus rule format. To load into the embedded Mimir:

1. Open Grafana at http://localhost:3001.
2. Navigate to **Alerting → Alert rules → Import**.
3. Upload each file from `deployments/observability/alerts/`.

For production Mimir/Alertmanager, copy the files to the ruler storage path or use the Mimir ruler API.

---

## Troubleshooting

| Symptom | Check |
|---------|-------|
| No MSSQL metrics in Prometheus | `docker compose logs otelcol` — look for `"receiver_type": "sqlserver"` at startup and any auth errors |
| Grafana data sources empty | Wait ~60s after `up`; if still empty check `docker compose logs lgtm` |
| `OTLP connection refused` | Ensure app OTLP endpoint points to `localhost:4317` (not to the internal `otelcol` sidecar port 55680) |
| `make observability-up` fails | Ensure Docker is running and ports 3001, 4317, 4318 are free |
