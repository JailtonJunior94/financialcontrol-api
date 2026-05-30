#!/usr/bin/env bash
# run_sqlcmd.sh — Executa queries SQL via sqlcmd usando as mesmas queries do codebase.
# Uso:
#   DATE=01/06/2026 ENVIRONMENT=Production bash scripts/run_sqlcmd.sh <query_name> [CATEGORY=<nome>]
#
# Queries disponíveis:
#   run_budget | run_budget_cards_and_others | run_budget_unified |
#   run_budget_full | run_balance | run_budget_category

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SQL_DIR="${SCRIPT_DIR}/sql"

# ------------------------------------------------------------------ #
# Helpers
# ------------------------------------------------------------------ #
die()  { echo "[ERROR] $*" >&2; exit 1; }
info() { echo "[INFO]  $*"; }

# ------------------------------------------------------------------ #
# Validações básicas
# ------------------------------------------------------------------ #
QUERY_NAME="${1:-}"
[[ -z "$QUERY_NAME" ]] && die "Informe o nome da query. Ex: bash run_sqlcmd.sh run_budget"

SQL_FILE="${SQL_DIR}/${QUERY_NAME}.sql"
[[ -f "$SQL_FILE" ]] || die "Arquivo SQL não encontrado: ${SQL_FILE}"

command -v sqlcmd >/dev/null 2>&1 || die "sqlcmd não encontrado. Instale o mssql-tools ou sqlcmd standalone."

[[ -z "${DATE:-}" ]] && die "Variável DATE não definida. Ex: DATE=01/06/2026"

# ------------------------------------------------------------------ #
# Carrega .env (ENVIRONMENT determina qual arquivo)
# ------------------------------------------------------------------ #
ENV_FILE="${REPO_ROOT}/.env"
[[ -f "$ENV_FILE" ]] || die "Arquivo .env não encontrado em: ${REPO_ROOT}"

# Exporta variáveis do .env, ignorando comentários e linhas vazias.
# Não sobrescreve variáveis já definidas no ambiente (env tem prioridade).
while IFS= read -r line || [[ -n "$line" ]]; do
    [[ "$line" =~ ^\s*# ]] && continue
    [[ -z "${line// }" ]]  && continue
    key="${line%%=*}"
    val="${line#*=}"
    # Só exporta se ainda não estiver no ambiente
    if [[ -z "${!key+x}" ]]; then
        export "$key"="$val"
    fi
done < "$ENV_FILE"

[[ -z "${MSSQL_CONNECTION_STRING:-}" ]] && die "MSSQL_CONNECTION_STRING não definida no .env ou no ambiente."

# ------------------------------------------------------------------ #
# Parse da connection string
# Formatos suportados:
#   sqlserver://user:pass@host
#   sqlserver://user:pass@host:port
#   sqlserver://user:pass@host?database=DB
#   sqlserver://user:pass@host:port?database=DB
# Senhas com '@' são suportadas (usa o último '@' como separador host/user)
# ------------------------------------------------------------------ #
conn="${MSSQL_CONNECTION_STRING}"

# Remove o prefixo 'sqlserver://'
conn="${conn#sqlserver://}"

# Extrai database (opcional — tudo após '?database=')
if [[ "$conn" == *"database="* ]]; then
    DB_NAME="${conn#*database=}"
    DB_NAME="${DB_NAME%%&*}"  # remove query params extras
else
    DB_NAME=""  # usa o database default do usuário no servidor
fi
conn="${conn%%\?*}"  # remove query string

# Extrai host (tudo após o último '@')
# Lida com senhas que contêm '@' pegando o último '@' como separador
HOST_PART="${conn##*@}"
USER_PART="${conn%@*}"  # tudo antes do último '@'

# Extrai porta, se presente (host:port)
SQL_HOST="${HOST_PART%%:*}"
SQL_PORT="${HOST_PART#*:}"
[[ "$SQL_PORT" == "$SQL_HOST" ]] && SQL_PORT="1433"  # sem porta explícita

SQL_USER="${USER_PART%%:*}"
SQL_PASS="${USER_PART#*:}"

[[ -z "$SQL_HOST" ]] && die "Não foi possível extrair o host da connection string."
[[ -z "$SQL_USER" ]] && die "Não foi possível extrair o usuário da connection string."
[[ -z "$SQL_PASS" ]] && die "Não foi possível extrair a senha da connection string."

# ------------------------------------------------------------------ #
# Converte DATE (DD/MM/YYYY) → YYYY-MM-DD para SQL Server
# ------------------------------------------------------------------ #
parse_date() {
    local input="$1"
    local day month year

    if [[ "$input" =~ ^([0-9]{2})/([0-9]{2})/([0-9]{4})$ ]]; then
        day="${BASH_REMATCH[1]}"
        month="${BASH_REMATCH[2]}"
        year="${BASH_REMATCH[3]}"
        # Normaliza para o primeiro dia do mês
        echo "${year}-${month}-01"
    elif [[ "$input" =~ ^([0-9]{4})-([0-9]{2})-([0-9]{2})$ ]]; then
        # Já está no formato ISO — normaliza o dia para 01
        echo "${BASH_REMATCH[1]}-${BASH_REMATCH[2]}-01"
    else
        die "Formato de DATE inválido: '${input}'. Use DD/MM/YYYY ou YYYY-MM-DD."
    fi
}

REF_DATE="$(parse_date "$DATE")"

# Calcula invoice date = refDate + 1 mês
add_one_month() {
    local iso_date="$1"
    local year="${iso_date:0:4}"
    local month="${iso_date:5:2}"
    local month_dec="${month#0}"  # remove zero à esquerda para aritmética

    month_dec=$((month_dec + 1))
    if [[ $month_dec -gt 12 ]]; then
        month_dec=1
        year=$((year + 1))
    fi

    printf "%04d-%02d-01" "$year" "$month_dec"
}

INVOICE_DATE="$(add_one_month "$REF_DATE")"

# ------------------------------------------------------------------ #
# Valida CATEGORY quando necessário
# ------------------------------------------------------------------ #
if [[ "$QUERY_NAME" == "run_budget_category" ]]; then
    [[ -z "${CATEGORY:-}" ]] && die "Variável CATEGORY não definida. Ex: CATEGORY=Alimentacao"
fi

# ------------------------------------------------------------------ #
# Monta parâmetros sqlcmd
# ------------------------------------------------------------------ #
CATEGORY_VAL="${CATEGORY:-}"

info "Conectando em ${SQL_HOST}:${SQL_PORT} | db: ${DB_NAME:-<default>} | usuário: ${SQL_USER}"
info "Data de referência : ${REF_DATE}"
info "Data das invoices  : ${INVOICE_DATE}"
[[ -n "$CATEGORY_VAL" ]] && info "Categoria          : ${CATEGORY_VAL}"
info "Query              : ${QUERY_NAME}"
echo ""

# sqlcmd flags:
#   -W  remove trailing spaces
#   -s  separador de colunas
#   -h  linhas de header por página (-1 = sem repetição)
#   -y / -Y  largura de colunas varchar/char (0 = sem truncamento)
sqlcmd \
    -S "${SQL_HOST},${SQL_PORT}" \
    -U "${SQL_USER}" \
    -P "${SQL_PASS}" \
    ${DB_NAME:+-d "$DB_NAME"} \
    -i "${SQL_FILE}" \
    -v date="${REF_DATE}" \
       invoiceDate="${INVOICE_DATE}" \
       category="${CATEGORY_VAL}" \
    -W \
    -s "|" \
    -h -1 \
    -y 0 \
    -Y 0
