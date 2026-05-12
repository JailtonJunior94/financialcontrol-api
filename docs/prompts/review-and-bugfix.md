# Prompt de Revisão e Bugfix de Código

Este prompt deve ser utilizado para orquestrar o fluxo de revisão de código seguido de correção de bugs (bugfix) quando necessário, garantindo conformidade com a governança do projeto e a qualidade técnica.

## 1. Revisão de Código (Skill review)

**Comando:** Use a skill `review` para revisar o diff atual.

**Contexto da Implementação:**
- **Tasks executadas:** `tasks/prd-modernization-unified-finance`
- **Skill usada na implementação:** `execute-task`
- **Áreas de risco:** `[performance, seguranca, contratos, concorrencia]`

**Focos Obrigatórios da Revisão:**
- **Corretude:** A implementação atende todos os Requisitos Funcionais (RFs) e critérios de aceite do PRD?
- **Regressão:** Alguma mudança quebra contrato público ou comportamento existente?
- **Segurança:** Há injeção de dependência insegura, dado sensível exposto ou validação faltando?
- **Testes:** Todos os cenários do critério de pronto estão cobertos?
- **Dívida Técnica:** O que precisará de refactor futuro?

**Saídas Esperadas (Veredito):**
- **Mandatório:** Não ter falsos positivos. Tenha certeza absoluta dos achados.
- **Formato:** Lista de achados por categoria: `Crítico`, `Importante`, `Sugestão`.
- **Detalhamento (Críticos):** Para cada achado crítico, informe: `arquivo:linha`, `descrição` e `correção sugerida`.
- **Veredito Final:** Escolha exatamente um: `APROVADO` / `APROVADO COM RESSALVAS` / `REPROVADO`.

---

## 2. Correção de Bugs (Skill bugfix)

**Gatilho:** Se o veredito for diferente de `APROVADO` ou houver bugs identificados.
**Comando:** Use a skill `bugfix` para corrigir TODOS os achados da revisão.

**Entrada para o Bugfix (da saída da skill review):**
- `[arquivo:linha]` Achado 1: `[descrição do problema]`
- `[arquivo:linha]` Achado 2: `[descrição do problema]`

**Comportamento Esperado após a Correção:**
- `[Descrever o comportamento correto para cada achado identificado]`

**Invariantes (Não podem mudar):**
- **Contratos Públicos:** Assinaturas de métodos, endpoints de API e tipos de DTOs públicos.
- **Tipos de Erro:** Erros canônicos que o restante do sistema depende para controle de fluxo.
- **Fluxos Adjacentes:** Comportamento de outros fluxos não afetados diretamente por estes achados.

---

## 3. Regras de Execução Não Negociáveis

1. **Análise de Causa Raiz:** Identifique a causa raiz de cada achado antes de escrever qualquer linha de código.
2. **Testes de Regressão:** Adicione ou corrija testes que provem que o bug foi eliminado e não voltará.
3. **Escopo Estrito:** Não altere comportamento fora do escopo dos achados listados.
4. **Validação Técnica (Go):** Ao finalizar, é obrigatório rodar e registrar o output de:
   - `gofmt -w .`
   - `go test ./...`
   - `golangci-lint run` (se disponível) ou `go vet ./...`
5. **Critério de Conclusão:** O bugfix só é considerado concluído com evidência de que a causa raiz foi eliminada e os testes passaram.

**Saídas Esperadas do Bugfix:**
1. **Diff:** Correção mínima necessária e idiomática.
2. **Testes:** Testes de regressão novos ou atualizados.
3. **Evidência:** Output dos comandos de validação (`go test`, `go vet`).
4. **Relatório:** Descrição da causa raiz e como a correção a endereça.
5. **Status Final:** O processo deve ser repetido até que o status seja verdadeiramente `APROVADO`.
