# GitHub Copilot CLI

Use `AGENTS.md` como instrucao principal deste repositorio.

## Instrucoes

1. Ler `AGENTS.md` no inicio da sessao.
2. `.agents/skills/` e a fonte de verdade dos fluxos procedurais.
3. `.github/agents/` sao wrappers leves que apontam para a habilidade correta.
4. Em tarefas de execucao, carregar apenas `AGENTS.md`, `agent-governance` e a skill operacional da linguagem ou atividade afetada.
5. Skills de planejamento (`analyze-project`, `create-prd`, `create-technical-specification`, `create-tasks`) entram apenas quando a tarefa pedir esse fluxo explicitamente.
6. Carregar referencias adicionais apenas quando a tarefa exigir.
7. Preservar estilo, arquitetura e fronteiras existentes antes de propor mudancas.
8. Validar mudancas com comandos proporcionais ao risco.

## Contrato de Carga Base

Antes de editar codigo, confirmar que estes arquivos foram lidos na sessao:

1. `AGENTS.md` — regras de arquitetura, modo de trabalho e restricoes.
2. `.agents/skills/agent-governance/SKILL.md` — governanca, DDD, erros, seguranca e testes sob demanda.
3. A skill de linguagem correspondente quando a tarefa alterar codigo:
   - Go: `.agents/skills/go-implementation/SKILL.md`
   - Node/TypeScript: `.agents/skills/node-implementation/SKILL.md`
   - Python: `.agents/skills/python-implementation/SKILL.md`

## Validacao

Ao concluir uma alteracao:

1. Rodar formatter nos arquivos alterados quando a stack oferecer.
2. Rodar testes direcionados aos modulos afetados.
3. Rodar lint quando disponivel e proporcional ao risco.
4. Registrar comandos executados e resultados de validacao.

## Stack

- Projeto com contexto Go detectado: carregar `.agents/skills/go-implementation/SKILL.md` ao alterar codigo Go.
- Validar a versao declarada em `go.mod` antes de introduzir APIs da linguagem ou novas dependencias.

## Orientacoes Especificas para Copilot

O GitHub Copilot suporta agents em `.github/agents/` e carrega `copilot-instructions.md` automaticamente, mas nao suporta hooks de enforcement. Para manter compliance:

1. Usar agents disponíveis em `.github/agents/` para delegar tarefas processuais (review, bugfix, execute-task, etc.).
2. Cada agent aponta para a skill canonica em `.agents/skills/` — seguir as etapas procedurais do SKILL.md referenciado.
3. Ao iniciar uma tarefa, confirmar que `AGENTS.md` e `agent-governance/SKILL.md` foram lidos.
4. Ao final da tarefa, executar os comandos de validacao descritos na secao Validacao acima.
5. Enforcement depende do modelo seguir as instrucoes — nao ha bloqueio automatico.
