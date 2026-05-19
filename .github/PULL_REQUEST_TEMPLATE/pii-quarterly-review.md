# PII Denylist — Auditoria Trimestral (RF-09.4)

Este PR e parte da auditoria trimestral obrigatoria da denylist de PII.
Mesmo sem alteracoes na lista, o PR deve ser aberto e mergeado para criar
evidencia auditavel para fins de conformidade LGPD.

## Checklist de revisao

- [ ] Revisao dos campos atuais da denylist em `internal/bootstrap/observability/redactor/denylist.go`
- [ ] Verificar se novos campos sensíveis foram introduzidos nos ultimos 3 meses (logs, spans, handlers, use cases)
- [ ] Confirmar que `TestPIIRedactedEndToEnd` continua cobrindo todos os campos da denylist
- [ ] Confirmar que o job `pii-guardrail` esta verde neste PR
- [ ] Adicionar entrada no `CHANGELOG.md` sob `[Unreleased]` → `Security`

## Resultado da revisao

Marque **uma** das opcoes abaixo:

- [ ] **Sem alteracoes** — lista revisada, nenhum campo novo identificado (`reviewed, no changes`)
- [ ] **Com alteracoes** — novos campos adicionados (listar abaixo)

### Campos adicionados (se houver)

<!-- Liste aqui os campos adicionados e o motivo -->

## Referencia

- Denylist: `internal/bootstrap/observability/redactor/denylist.go`
- Guardrail CI: job `pii-guardrail` em `.github/workflows/ci-cd.yml`
- Proxima auditoria: 12 semanas apos o merge deste PR
