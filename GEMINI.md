# Gemini CLI

Use `AGENTS.md` como fonte canonica das regras deste repositorio.

## Instrucoes

1. Ler `AGENTS.md` no inicio da sessao.
2. `.agents/skills/` e a fonte de verdade dos fluxos procedurais.
3. `.gemini/commands/` sao adaptadores finos que apontam para a habilidade correta.
4. Em tarefas de execucao, carregar apenas `AGENTS.md`, `agent-governance` e a skill operacional da linguagem ou atividade afetada.
5. Skills de planejamento (`analyze-project`, `create-prd`, `create-technical-specification`, `create-tasks`) entram apenas quando a tarefa pedir esse fluxo explicitamente.
6. Carregar referencias adicionais apenas quando a tarefa exigir.
7. Preservar estilo, arquitetura e fronteiras existentes antes de propor mudancas.
8. Validar mudancas com comandos proporcionais ao risco.

## Stack

- Projeto com contexto Go detectado: carregar `.agents/skills/go-implementation/SKILL.md` ao alterar codigo Go.
- Validar a versao declarada em `go.mod` antes de introduzir APIs da linguagem ou novas dependencias.


## Orientacoes Especificas para Gemini

O Gemini CLI nao suporta hooks, agents ou rules nativos. Para modelar o fluxo de governanca:

1. Ao iniciar uma tarefa, ler `AGENTS.md` e `.agents/skills/agent-governance/SKILL.md` como contexto base antes de editar codigo.
2. Usar `@workspace.<command>` para invocar o wrapper TOML correspondente e evitar colisao com comandos nativos das skills.
3. Seguir as etapas procedurais do SKILL.md carregado pelo comando como se fossem instrucoes sequenciais.
4. Ao final da tarefa, executar os comandos de validacao descritos na secao Validacao do `AGENTS.md`.
5. Nao confiar em enforcement automatico — a compliance depende de seguir as instrucoes procedurais manualmente.
