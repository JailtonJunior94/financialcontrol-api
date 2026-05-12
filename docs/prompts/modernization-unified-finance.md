# Enriched Prompt: Refactor e Modernização dos Módulos Financeiros (Unified Finance Module)

## Objetivo
Refatorar, modernizar e unificar os módulos `internal/modules/billing`, `internal/modules/transactions` e `internal/modules/invoicing` em um único módulo coeso. O objetivo é permitir a gestão unificada de transações do dia a dia (entradas e saídas), compras parceladas em cartões de crédito e faturamento mensal, utilizando uma arquitetura robusta e escalável.

## Requisitos Mandatórios (Não Negociáveis)
1. **Arquitetura & Design:**
   - **Modular Monolith:** Respeitar fronteiras rígidas entre módulos. O módulo `finance` não deve importar pacotes internos de `cards` ou `categories`.
   - **Comunicação Inter-modular:** Para obter dados de Cartões ou Categorias, utilizar o padrão **Ports & Adapters**. O módulo `finance` deve definir interfaces (ex: `CardProvider`, `CategoryProvider`) que serão implementadas pelos respectivos módulos e injetadas via DI.
   - **Clean Architecture:** Separação clara entre Domain, Application, Infrastructure e Interfaces (HTTP).
   - **Domain-Driven Design (DDD):** Identificação de Bounded Contexts, Agregados (Aggregate Roots), Entidades e Value Objects. Definir projeções locais de Cartão e Categoria no domínio de `finance`.
   - **Object Calisthenics:** Aplicação rigorosa das 9 regras (ex: apenas um nível de indentação por método, sem `else`, wrap de tipos primitivos, etc.).
   - **Padrões de Projeto:** Uso de Factory, Repository, Strategy e outros padrões Go-idiomáticos onde aplicável.

2. **Stack Técnica:**
   - **Linguagem:** Go (respeitando a versão no `go.mod`).
   - **Shared Packages:** Utilizar obrigatoriamente as bases em `pkg/entity`, `pkg/shared`, `pkg/customerrors` e `pkg/database`.

3. **Funcionalidades & Regras de Negócio:**
   - **Transações Unificadas:** Um único CRUD para gerenciar entradas (ex: salário via PIX) e saídas (ex: pagamento de boleto).
   - **Compras Parceladas:** Capacidade de lançar uma compra (ex: viagem em 10x) categorizada (ex: Metas e Lazer).
   - **Gestão de Faturas (Credit Card Invoicing):**
     - Referenciar o mês e o valor total da fatura na tabela principal de transações.
     - Tabela de faturas (`invoices`) e itens de fatura (`invoice_items`) sem necessidade de processo manual de fechamento.
     - **Regra de Vencimento:** O sistema deve determinar automaticamente em qual fatura lançar a compra com base na data da transação e nas regras de fechamento/vencimento do cartão.
   - **Categorização:** Suporte a múltiplas categorias (ex: Metas, Lazer, Contas Fixas).

## Estrutura de Saída Esperada
O agente deve fornecer um plano detalhado e a implementação seguindo esta estrutura:

1. **Análise de Domínio:**
   - Definição do Agregado Principal (ex: `Transaction` ou `FinanceEntry`).
   - Mapeamento de Value Objects (ex: `Amount`, `Installment`, `Category`).
   - Definição de Domain Services para regras complexas (ex: Cálculo de Faturas).

2. **Especificação Técnica:**
   - Definição das interfaces (Ports) no domínio.
   - Estrutura de pastas unificada em `internal/modules/finance` (ou nome sugerido).
   - Esquema de banco de dados (SQL ou GORM tags) que suporte a unificação.

3. **Implementação de Exemplo:**
   - Código Go das entidades de domínio com validações.
   - Exemplo de um Use Case (Application Service) para "Lançar Compra Parcelada".

## Critérios de Aceitação
- O código deve ser 100% livre de `else` (Object Calisthenics).
- Não deve haver lógica de infraestrutura (SQL, HTTP) misturada com lógica de negócio.
- A unificação deve reduzir a duplicidade de lógica de "Totalização" observada nos módulos antigos.
- Testes unitários devem ser previstos para as regras de cálculo de parcelas e faturas.

## Descomissionamento e Limpeza (Mandatório)
Após a implementação, validação e migração bem-sucedida para o novo módulo unificado:
- **Exclusão de Módulos Antigos:** Remover permanentemente as pastas `internal/modules/billing`, `internal/modules/transactions` e `internal/modules/invoicing`.
- **Limpeza de Referências:** Atualizar `internal/modules/registry.go` e quaisquer outros arquivos de bootstrap/main para remover imports e registros dos módulos excluídos.
- **Remoção de Testes Órfãos:** Garantir que testes relacionados aos módulos antigos também sejam removidos.
