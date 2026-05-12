#!/usr/bin/env bash
# Hook de pre-execucao para Gemini CLI: bloqueia execucao sem preload de governanca.
#
# Uso: configurar no GEMINI.md ou via --hook do Gemini CLI:
#   gemini --hook "bash .gemini/hooks/validate-preload.sh"
#
# Unlock: exportar GOVERNANCE_PRELOAD_CONFIRMED=1 na sessao atual.
# Consulte GEMINI.md para instrucoes completas de preload.

set -euo pipefail

if [[ "${GOVERNANCE_PRELOAD_CONFIRMED:-}" != "1" ]]; then
  echo "ERRO: governanca nao carregada." >&2
  echo "Antes de editar codigo, confirme que foram lidos:" >&2
  echo "  - AGENTS.md" >&2
  echo "  - .agents/skills/agent-governance/SKILL.md" >&2
  echo "  - .agents/skills/<linguagem>-implementation/SKILL.md (linguagem afetada)" >&2
  echo "Execute: export GOVERNANCE_PRELOAD_CONFIRMED=1" >&2
  echo "Consulte GEMINI.md para instrucoes completas." >&2
  exit 1
fi
