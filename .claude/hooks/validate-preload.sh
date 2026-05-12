#!/usr/bin/env bash
# Hook PreToolUse opcional: verifica se o contrato de carga base foi cumprido
# antes de permitir edicoes em codigo.
#
# Para habilitar, adicione ao .claude/settings.local.json:
#
#   "PreToolUse": [{
#     "matcher": "Edit|Write",
#     "hooks": [{"type": "command", "command": "bash .claude/hooks/validate-preload.sh"}]
#   }]
#
# Este hook bloqueia por padrao. Use GOVERNANCE_PRELOAD_MODE=warn para desabilitar o bloqueio.
# Entrada: JSON do tool use via stdin.
#
# Modos (via variavel de ambiente GOVERNANCE_PRELOAD_MODE):
#   fail  — emite lembrete em stderr, exit 1 (bloqueia a edicao) [DEFAULT]
#   warn  — emite lembrete em stderr, exit 0 (nao bloqueia, opt-out explícito)
#
# Unlock (override do bloqueio sem mudar o modo):
#   GOVERNANCE_PRELOAD_CONFIRMED=1  — bypass do bloqueio para sessoes que ja
#                                     confirmaram o contrato. Util em ferramentas
#                                     single-round (Codex, Copilot, Gemini CLI).
#
# Gate adicional por linguagem (Etapa 2.4 de execute-task):
#   GOVERNANCE_LANG_LOADED  — CSV das linguagens cuja SKILL de implementacao foi
#                             lida na sessao. Mapping extensao -> linguagem:
#                               .go            -> go
#                               .py            -> python
#                               .ts/.tsx/.js/.jsx -> node
#                             Se a linguagem do arquivo editado nao estiver na lista,
#                             o hook bloqueia (fail) ou avisa (warn). Bypass com
#                             GOVERNANCE_LANG_CONFIRMED=1 para sessoes single-round.

set -euo pipefail

GOVERNANCE_PRELOAD_MODE="${GOVERNANCE_PRELOAD_MODE:-fail}"
GOVERNANCE_PRELOAD_CONFIRMED="${GOVERNANCE_PRELOAD_CONFIRMED:-0}"
GOVERNANCE_LANG_LOADED="${GOVERNANCE_LANG_LOADED:-}"
GOVERNANCE_LANG_CONFIRMED="${GOVERNANCE_LANG_CONFIRMED:-0}"

HOOK_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=../../scripts/lib/parse-hook-input.sh
source "$HOOK_DIR/../../scripts/lib/parse-hook-input.sh" 2>/dev/null \
  || source "$(cd "$HOOK_DIR/../.." && pwd)/scripts/lib/parse-hook-input.sh" 2>/dev/null \
  || { echo "AVISO: parse-hook-input.sh nao encontrado" >&2; exit 0; }

_stdin="$(cat)"
file_path="$(printf '%s' "$_stdin" | parse_file_path)"

[[ -n "$file_path" ]] || exit 0

# Mapeia extensao do arquivo para a linguagem cuja SKILL deve estar carregada.
# Retorna string vazia se a extensao nao for de codigo de producao rastreado.
detect_lang() {
  case "$1" in
    *.go)                       echo "go" ;;
    *.py)                       echo "python" ;;
    *.ts|*.tsx|*.js|*.jsx)      echo "node" ;;
    *)                          echo "" ;;
  esac
}

lang="$(detect_lang "$file_path")"

# Arquivo nao-codigo: passa direto.
[[ -z "$lang" ]] && exit 0

echo "LEMBRETE: antes de editar codigo, confirme que AGENTS.md e agent-governance/SKILL.md foram lidos nesta sessao." >&2

# Gate 1: contrato de carga base (governance).
if [[ "$GOVERNANCE_PRELOAD_CONFIRMED" != "1" ]]; then
  if [[ "$GOVERNANCE_PRELOAD_MODE" == "fail" ]]; then
    echo "GOVERNANCE_PRELOAD_MODE=fail: bloqueando edicao ate que contrato de carga seja confirmado." >&2
    echo "Para prosseguir: export GOVERNANCE_PRELOAD_CONFIRMED=1" >&2
    exit 1
  fi
fi

# Gate 2: skill da linguagem afetada (Etapa 2.4 de execute-task).
# Bypass: GOVERNANCE_LANG_CONFIRMED=1 (single-round) OU lang em GOVERNANCE_LANG_LOADED (CSV).
if [[ "$GOVERNANCE_LANG_CONFIRMED" != "1" ]]; then
  # Procura $lang na CSV (com ou sem espacos), em qualquer posicao.
  if ! printf '%s' ",$GOVERNANCE_LANG_LOADED," | tr -d '[:space:]' | grep -Fq ",$lang,"; then
    echo "LEMBRETE: editando arquivo $lang, mas .agents/skills/$lang-implementation/SKILL.md nao consta como carregada." >&2
    if [[ "$GOVERNANCE_PRELOAD_MODE" == "fail" ]]; then
      echo "GOVERNANCE_PRELOAD_MODE=fail: bloqueando edicao ate que a skill de linguagem seja carregada." >&2
      echo "Para prosseguir: export GOVERNANCE_LANG_LOADED=\"\${GOVERNANCE_LANG_LOADED:+\$GOVERNANCE_LANG_LOADED,}$lang\"" >&2
      echo "Ou, em single-round: export GOVERNANCE_LANG_CONFIRMED=1" >&2
      exit 1
    fi
  fi
fi

exit 0
