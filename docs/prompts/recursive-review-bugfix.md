# Prompt: Fluxo de Revisão e Correção Recursiva (Review & Bugfix)

Este prompt foi projetado para execução contínua até que o status **APPROVED** seja alcançado, garantindo a máxima qualidade e segurança na modernização de observabilidade.

## Contexto da Implementação
- **Tasks:** `tasks/prd-modernization-observability`
- **Skill base:** `execute-task`
- **Stack:** Go (Monolito), Fiber
- **Áreas de Risco:** Performance, Segurança (incluindo PII Denylist RF-27), Contratos Públicos, Concorrência.

---

## 🔄 Rodada de Execução

### ETAPA 1: Revisão Crítica (Skill: review)
Execute a skill `review` no diff atual com rigor máximo. **Não seja preguiçoso.**

**Focos Obrigatórios:**
1. **Corretude:** Atendimento total aos RFs e critérios de aceite do PRD.
2. **Regressão:** Mudanças em contratos públicos, assinaturas ou tipos que quebrem compatibilidade.
3. **Segurança:** 
   - Verifique a Denylist de PII (RF-27): `pan`, `cpf`, `cnpj`, `email`, `token`, etc.
   - Injeção de dependência insegura ou validações ausentes.
4. **Concorrência:** Race conditions, vazamento de goroutines ou uso inseguro de maps/slices em Go.
5. **Testes:** Cobertura de cenários felizes e de erro (Edge cases).
6. **Dívida Técnica:** Identificação de refactors necessários.

**Saída Obrigatória da Revisão:**
- **Veredito Final:** [APPROVED | APPROVED WITH RESERVATIONS | REPROVED]
- **Lista de Achados:** Categorizados por [Crítico], [Importante], [Sugestão].
- **Para Achados Críticos:** `[arquivo:linha]` | Descrição detalhada | Correção sugerida.

> **Regra de Ouro:** Se houver QUALQUER achado Crítico ou Importante, o veredito deve ser **REPROVED**. O status **APPROVED** só é permitido com ZERO achados críticos/importantes.

---

### ETAPA 2: Correção de Achados (Skill: bugfix)
Se o veredito for diferente de **APPROVED**, execute a skill `bugfix` para cada achado listado.

**Diretrizes de Correção:**
1. **Causa Raiz:** Identifique a causa real antes de codificar.
2. **Testes de Regressão:** Adicione ou corrija testes que provem que o erro foi eliminado e não voltará.
3. **Invariantes:** Não altere contratos públicos ou comportamentos fora do escopo do bug.
4. **Validação Proporcional:** 
   - Execute `gofmt -w .`
   - Execute `go test ./...`
   - Execute `golangci-lint run`

**Saída Obrigatória do Bugfix:**
- Diff com correção mínima e segura.
- Evidência de validação (Output dos testes e lint).
- Descrição da causa raiz eliminada.

---

## 🏁 Critério de Parada
- Repita o ciclo (Revisão -> Correção) até que a Etapa 1 retorne o veredito **APPROVED**.
- **Segurança contra Loops:** Se a profundidade de invocação (`AI_INVOCATION_DEPTH`) atingir 2 rodadas sem convergência, pare e solicite intervenção manual detalhando os impedimentos.

## 🚫 Restrições Não Negociáveis
- **Falsos Positivos:** Avalie cada achado com base em evidência no código, não em suposições.
- **Preguiça:** Revisões superficiais serão consideradas falha na tarefa.
- **PII:** Qualquer exposição de campo da Denylist sem redator é erro CRÍTICO.
