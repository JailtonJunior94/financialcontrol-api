# migration — Runbook Operacional

Binário CLI dedicado à aplicação de migrações de schema MSSQL. Executado como **Init Container** ou **Job** no Kubernetes antes da subida da API, garantindo que o schema esteja atualizado de forma determinística e auditável.

## 1. Visão Geral

O binário `migration` aplica todos os arquivos `.sql` em `pkg/database/migrations/` (embutidos via `go:embed`) ao banco MSSQL configurado. Usa o motor `golang-migrate` (via `devkit-go/pkg/database/migration`) com driver `sqlserver`.

- **Modo único:** `apply-all` (forward-only; sem subcomandos).
- **Idempotente:** re-executar contra banco já atualizado retorna exit 0 sem reaplicar scripts.
- **Standalone:** não depende da API em execução; adequado para Init Container ou Job K8s.

```bash
# Build local
make migrate-build

# Aplicar migrações pendentes
make migrate-up

# Baseline em ambientes preexistentes (ver seção 6)
make migrate-baseline BASELINE=1
```

## 2. Variáveis de Ambiente

| Variável | Obrigatória | Default | Descrição |
|---|---|---|---|
| `MSSQL_CONNECTION_STRING` | **sim** | — | DSN do MSSQL. Deve começar com `sqlserver://`. Aceita query params TLS (ver seção 8). |
| `MIGRATION_TIMEOUT` | não | `10m` | Timeout global da execução. Excedido, encerra com exit 1. |
| `MIGRATION_BASELINE` | não | — | Versão uint (≥ 1) para operação de baseline (ver seção 6). Aceita qualquer N ≥ 1. `0` é rejeitado. |
| `CI_BUILD_ID` | não | — | ID do build CI, propagado nos logs de auditoria. |
| `GITHUB_RUN_ID` | não | — | ID do run GitHub Actions, propagado nos logs de auditoria. |
| `K8S_POD_NAME` | não | — | Nome do pod K8s, propagado nos logs de auditoria. |
| `HOSTNAME` | não | — | Fallback quando `K8S_POD_NAME` ausente. |
| `K8S_NAMESPACE` | não | — | Namespace K8s, propagado nos logs de auditoria. |
| `TRIGGERED_BY` | não | — | Operador/triggering principal, propagado nos logs de auditoria. |
| `GIT_COMMIT_SHA` | não | — | SHA do commit, propagado nos logs de auditoria. |

**Formato do DSN (`MSSQL_CONNECTION_STRING`):**

```
sqlserver://user:password@host:port?database=FinancialControlDB&encrypt=true&trustServerCertificate=false
```

## 3. Exit Codes

| Código | Significado |
|---|---|
| `0` | Sucesso — migrações aplicadas, sem mudanças (already up-to-date) ou pasta vazia (no-op). |
| `1` | Qualquer falha — erro de conexão, sintaxe SQL, checksum mismatch, ordering inválido, timeout. |

Padrões de log para distinguir categorias:

| Mensagem JSON (`msg`) | Nível | Significado |
|---|---|---|
| `migration up complete` | `INFO` | Todas as migrações aplicadas com sucesso. |
| `migration up no change` | `INFO` | Banco já atualizado; sem novas migrações. Exit 0. |
| `migration up failed` | `ERROR` | Falha na aplicação. Detalhes em `error`, `version`, `dirty`. |
| `migrations dir empty` | `WARN` | Nenhum arquivo `.sql` encontrado. Exit 0 (no-op). |
| `connection failed, will retry` | `WARN` | Tentativa de conexão falhou; backoff exponencial em curso. |
| `migration starting` | `INFO` | Início da execução com build info e auditoria. |
| `preflight ok` | `INFO` | Pre-flight `SELECT 1` bem-sucedido antes do lock. |
| `migration shutdown` | `INFO` | Shutdown gracioso após SIGTERM/SIGINT. |

## 4. Logs Estruturados

O CLI emite logs em **JSON estruturado** no `stdout`, compatíveis com Loki, Elastic e Datadog.

**Campos canônicos:**

| Campo | Tipo | Descrição |
|---|---|---|
| `time` | string (RFC3339) | Timestamp da entrada. |
| `level` | string | `INFO`, `WARN` ou `ERROR`. |
| `msg` | string | Mensagem legível (ver tabela seção 3). |
| `database` | string | Nome do banco (sem credenciais). |
| `version` | int | Versão da migração corrente. |
| `dirty` | bool | Indica dirty state na tabela de controle. |
| `duration` | string | Duração da operação completa. |
| `error` | string | Mensagem de erro (somente em falha). |
| `attempt` | int | Número da tentativa de conexão (retry). |
| `delay` | string | Delay do próximo retry. |
| `binary.version` | string | Versão do binário (injetada via `-ldflags`). |
| `binary.commit` | string | SHA do commit de build. |
| `binary.build_date` | string | Data de build. |
| `build_id` | string | `CI_BUILD_ID` ou `GITHUB_RUN_ID`. |
| `pod_name` | string | `K8S_POD_NAME` ou `HOSTNAME`. |
| `triggered_by` | string | `TRIGGERED_BY`. |

**Redação de credenciais:** senha, DSN completo e tokens são substituídos por `***` em todos os logs. Host, database e usuário aparecem apenas para diagnóstico.

**Exemplo de log de sucesso:**

```json
{"time":"2026-05-09T10:00:00Z","level":"INFO","msg":"migration up complete","database":"FinancialControlDB","version":1,"dirty":false,"duration":"1.23s","binary.version":"v1.0.0","binary.commit":"abc1234","binary.build_date":"2026-05-09"}
```

## 5. Troubleshooting

**Q: O CLI retorna exit 1 com `connection retry exhausted`.**

Causa: banco MSSQL inacessível após 8 tentativas (~30s de backoff exponencial, conforme RNF-05).

Ação:
1. Verificar `MSSQL_CONNECTION_STRING` — DSN correto, host e porta acessíveis.
2. Verificar que o servidor MSSQL está rodando (`SELECT 1` manual).
3. Em K8s: aumentar `initialDelaySeconds` do Init Container para dar tempo ao banco subir.
4. Ver campo `attempt` e `error` nos logs JSON para diagnóstico detalhado.

---

**Q: O CLI retorna exit 1 com `context deadline exceeded` / timeout (RNF-10).**

Causa: `MIGRATION_TIMEOUT` expirou antes do término das migrações.

Ação:
1. Aumentar `MIGRATION_TIMEOUT` (ex.: `MIGRATION_TIMEOUT=30m` para destrutivas longas).
2. Verificar se há lock contention (`sp_getapplock`) de outra instância concorrente.
3. Rodar em janela off-peak (D-27) para migrações destrutivas.

---

**Q: Falha de autenticação / credential mismatch (RNF-06).**

Causa: usuário/senha incorretos na DSN.

Ação:
1. Verificar `MSSQL_CONNECTION_STRING` — usuário e senha corretos.
2. Confirmar que as credenciais têm permissão DDL no banco alvo.
3. Credenciais **nunca aparecem em logs**; inspecionar a env diretamente se suspeitar de problema de configuração.

---

**Q: O CLI retorna exit 1 com `dirty` state (D-60).**

Causa: migração anterior falhou no meio da execução, deixando a tabela `dbo.schema_migrations` com `dirty=true`.

Ação (recovery via baseline — reusa D-51):

```bash
# Identificar a versão que ficou dirty nos logs (campo "version")
# Forçar a versão N para limpar o dirty state
MSSQL_CONNECTION_STRING="..." MIGRATION_BASELINE=N make migrate-baseline

# Após baseline, executar Up normalmente
make migrate-up
```

---

**Q: `pre-flight check` falhou (RNF-14).**

Causa: conexão estabelecida mas banco não respondeu ao `SELECT 1` — conexão degradada ou banco sobrecarregado.

Ação:
1. Verificar health do servidor MSSQL.
2. Aguardar e re-executar; timeout global ainda se aplica.

---

**Q: Ambiente preexistente com tabelas já criadas (baseline em banco pré-existente — D-42).**

Ver seção 6 para o procedimento completo.

## 6. Procedimento de Baseline (Ambientes Preexistentes)

Use quando o banco já possui o schema inicial legado (ex.: `dbo.User`, `dbo.Card`, `dbo.Invoice`) mas **não** tem a tabela `dbo.schema_migrations` do golang-migrate.

**Passo a passo:**

1. **Identificar a versão atual do schema.** Se o schema inicial completo da versão `000001` já existe, a versão de baseline é `1`.

2. **Executar baseline uma única vez:**

```bash
MSSQL_CONNECTION_STRING="sqlserver://user:pass@host:1433?database=DB" \
  MIGRATION_BASELINE=1 \
  make migrate-baseline
```

Isso registra a versão N em `dbo.schema_migrations` sem re-executar os scripts DDL, marcando essas migrações como já aplicadas.

3. **Executar Up normalmente em todas as execuções seguintes:**

```bash
make migrate-up
```

A partir deste ponto, o CLI aplica apenas migrações com versão > N.

**Importante:** `MIGRATION_BASELINE=0` é rejeitado com erro imediato. Versões válidas: `>= 1`.

## 7. Rollback Manual (D-28)

O CLI é **forward-only** (D-1). Rollback é via restore de snapshot do banco:

1. **Parar a API** — garantir que nenhuma escrita ocorre durante o restore.

2. **Restaurar snapshot** pré-migration (D-17). O snapshot deve ter sido criado antes do Job de migração em prod.

3. **Reconciliar `dbo.schema_migrations`** pós-restore:
   - Se o snapshot for anterior à criação da tabela de controle: use o procedimento de baseline (seção 6) para re-registrar a versão correspondente ao snapshot.
   - Se o snapshot incluir a tabela de controle: a tabela reflete o estado pré-migration automaticamente.

4. **Validar integridade:**

```sql
SELECT version, dirty FROM dbo.schema_migrations;
-- dirty deve ser 0 (false); version deve corresponder ao snapshot
```

5. **Re-executar o baseline se necessário** (dirty=true após restore):

```bash
MIGRATION_BASELINE=<versao_do_snapshot> make migrate-baseline
```

6. **Subir a API** — schema e tabela de controle estão alinhados.

## 8. Recomendações de TLS em Produção (D-58, D-32)

A configuração de TLS é controlada exclusivamente pela DSN. Em produção, **sempre** incluir os parâmetros:

```
sqlserver://user:pass@host:1433?database=DB&encrypt=true&trustServerCertificate=false
```

| Parâmetro | Valor recomendado em prod | Descrição |
|---|---|---|
| `encrypt` | `true` | Habilita TLS para toda a comunicação com o banco. |
| `trustServerCertificate` | `false` | Valida o certificado do servidor; rejeita self-signed não confiáveis. |

Sem esses parâmetros, a conexão trafega em texto claro — risco aceito conscientemente para ambientes não-prod (D-32). Revisitar em hardening futuro de segurança para tornar obrigatório via validação no bootstrap.

## 9. Janela Operacional (D-27)

- **Migrações aditivas** (CREATE TABLE, ADD COLUMN): podem rodar a qualquer momento.
- **Migrações destrutivas** (DROP, ALTER COLUMN, TRUNCATE, rebuild de índice): requerem **janela off-peak** recomendada e **revisão obrigatória de DBA/tech lead** via PR (D-19, D-25).
- **Em produção:** agendar o Job K8s fora do horário de pico; coordenar com o time de SRE.

## 10. SLA / Target (D-36)

| Categoria | p95 target |
|---|---|
| Migrações típicas (DDL aditivo leve) | ≤ 2 minutos |
| Migrações destrutivas (DROP/ALTER/rebuild de índice) | ≤ 30 minutos |

Os targets são referência para configuração de `MIGRATION_TIMEOUT` e alertas no pipeline. Podem ser calibrados após primeiras execuções em produção. Sem gate técnico além do timeout global (RNF-10).

## 11. Rollout Faseado (D-16)

Sequência obrigatória antes de cada release com migrações:

```
dev → staging → prod
```

1. **dev:** rodar baseline (se necessário) + `migrate-up`; smoke test da API com novos schemas.
2. **staging:** idem dev; validar performance sob carga; simular rolling deploy para testar lock advisory (RNF-09).
3. **prod:**
   - Criar **snapshot do banco** (D-17) — pré-requisito obrigatório.
   - Agendar Job em **janela off-peak** (D-27) para destrutivas.
   - Executar Job de migração; monitorar logs JSON em Loki.
   - Fazer rolling restart da API.
   - Em caso de falha: rollback via snapshot (seção 7).
