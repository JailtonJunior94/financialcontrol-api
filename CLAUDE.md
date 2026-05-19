# Claude Code

Use `AGENTS.md` como fonte canonica das regras deste repositorio.

## Instrucoes

1. Ler `AGENTS.md` no inicio da sessao.
2. `.claude/skills/` sao symlinks para `.agents/skills/` — a fonte de verdade e sempre `.agents/skills/`.
3. `.claude/agents/` sao wrappers leves que delegam para a habilidade canonica.
4. Em tarefas de execucao, carregar apenas `AGENTS.md`, `agent-governance` e a skill operacional da linguagem ou atividade afetada.
5. Skills de planejamento (`analyze-project`, `create-prd`, `create-technical-specification`, `create-tasks`) entram apenas quando a tarefa pedir esse fluxo explicitamente.
6. Carregar referencias adicionais apenas quando a tarefa exigir.
7. Preservar estilo, arquitetura e fronteiras existentes antes de propor mudancas.
8. Validar mudancas com comandos proporcionais ao risco.

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

## Stack

- Projeto com contexto Go detectado: carregar `.agents/skills/go-implementation/SKILL.md` ao alterar codigo Go.
- Validar a versao declarada em `go.mod` antes de introduzir APIs da linguagem ou novas dependencias.
