# Relatório Financeiro — Junho 2026

> **Orçamento mensal: R$ 13.874,40** · Dados extraídos via `MSSQL_CONNECTION_STRING`

---

## Resumo do Mês

| Categoria            | %   | Planejado       | Gasto Cartão  | Gasto Outros  | Total Gasto   | Ainda Posso Gastar |
|----------------------|-----|-----------------|---------------|---------------|---------------|--------------------|
| Custos fixos         | 40% | R$ 5.549,76     | R$ 284,14     | R$ 1.166,51   | R$ 1.450,65   | **R$ 4.099,11**    |
| Conforto             | 10% | R$ 1.387,44     | R$ 34,90      | R$ 400,00     | R$ 434,90     | **R$ 952,54**      |
| Metas                | 15% | R$ 2.081,16     | R$ 1.664,74   | R$ 0,00       | R$ 1.664,74   | **R$ 416,42**      |
| Prazeres             | 10% | R$ 1.387,44     | R$ 1,00       | R$ 0,00       | R$ 1,00       | **R$ 1.386,44**    |
| Conhecimento         | 5%  | R$ 693,72       | R$ 392,82     | R$ 0,00       | R$ 392,82     | **R$ 300,91**      |
| Liberdade Financeira | 20% | R$ 2.774,88     | R$ 0,00       | R$ 2.774,88   | R$ 2.774,88   | **R$ 0,00**        |
| **TOTAL**            |     | **R$ 13.874,40** | **R$ 2.377,60** | **R$ 4.341,39** | **R$ 6.718,99** | **R$ 7.155,42** |

---

## 💳 Fatura do Cartão — O que Separar

> Valor exato a reservar para pagar a fatura do cartão de crédito:

| Categoria    | Subcategoria | Valor no Cartão |
|--------------|--------------|-----------------|
| Custos fixos | Serviços     | R$ 284,14       |
| Conforto     | Streamings   | R$ 34,90        |
| Metas        | Lazer        | R$ 1.664,74     |
| Prazeres     | Eletrônicos  | R$ 1,00         |
| Conhecimento | Academia     | R$ 75,00        |
| Conhecimento | Educação     | R$ 317,82       |
| **Total fatura** |          | **R$ 2.377,60** |

---

## 📊 Saldo Disponível por Categoria

### Custos fixos — R$ 4.099,11 disponível

| Origem                | Valor         |
|-----------------------|---------------|
| Cartão (Serviços)     | R$ 284,14     |
| Boletos/Contas        | R$ 1.166,51   |
| **Total gasto**       | **R$ 1.450,65** |
| Planejado             | R$ 5.549,76   |
| **Ainda posso gastar** | **R$ 4.099,11** |

### Conforto — R$ 952,54 disponível

| Origem                  | Valor         |
|-------------------------|---------------|
| Cartão (Streamings)     | R$ 34,90      |
| Boletos/Contas (Faxina) | R$ 400,00     |
| **Total gasto**         | **R$ 434,90** |
| Planejado               | R$ 1.387,44   |
| **Ainda posso gastar**  | **R$ 952,54** |

### Metas — R$ 416,42 disponível

| Origem                | Valor         |
|-----------------------|---------------|
| Cartão (Lazer)        | R$ 1.664,74   |
| Boletos/Contas        | R$ 0,00       |
| **Total gasto**       | **R$ 1.664,74** |
| Planejado             | R$ 2.081,16   |
| **Ainda posso gastar** | **R$ 416,42** |

### Prazeres — R$ 1.386,44 disponível

| Origem                | Valor         |
|-----------------------|---------------|
| Cartão (Eletrônicos)  | R$ 1,00       |
| Boletos/Contas        | R$ 0,00       |
| **Total gasto**       | **R$ 1,00**   |
| Planejado             | R$ 1.387,44   |
| **Ainda posso gastar** | **R$ 1.386,44** |

### Conhecimento — R$ 300,91 disponível

| Subcategoria | Valor no Cartão |
|--------------|-----------------|
| Academia     | R$ 75,00        |
| Educação     | R$ 317,82       |
| **Total**    | **R$ 392,82**   |

| Planejado   | Total Gasto | Ainda Posso Gastar |
|-------------|-------------|---------------------|
| R$ 693,72   | R$ 392,82   | **R$ 300,91**       |

### Liberdade Financeira — R$ 0,00 de saldo (100% comprometido)

| Origem          | Valor         |
|-----------------|---------------|
| Gasto fixo      | R$ 2.774,88   |
| Planejado       | R$ 2.774,88   |
| **Saldo**       | **R$ 0,00**   |

---

## 🔄 Como atualizar este relatório

```bash
# Conexão lida de MSSQL_CONNECTION_STRING no .env
DATE=01/06/2026 ENVIRONMENT=Production make query_budget_unified
DATE=01/06/2026 ENVIRONMENT=Production make query_balance
DATE=01/06/2026 ENVIRONMENT=Production make query_budget_cards_and_others
```

---

## Referência de Comandos

| Comando                              | Parâmetros                        | O que faz                                              |
|--------------------------------------|-----------------------------------|--------------------------------------------------------|
| `make run_sync`                      | `ENVIRONMENT`                     | Sincroniza transações com boletos                      |
| `make run_budget`                    | `ENVIRONMENT`, `DATE`             | Orçamento por categoria                                |
| `make run_budget_cards_and_others`   | `ENVIRONMENT`, `DATE`             | Gasto por cartão vs. boletos por categoria             |
| `make run_budget_unified`            | `ENVIRONMENT`, `DATE`             | Visão completa: planejado, cartão, outros, saldo       |
| `make run_budget_full`               | `ENVIRONMENT`, `DATE`             | Budget + Cards em sequência                            |
| `make run_balance`                   | `ENVIRONMENT`, `DATE`             | Saldo real + detalhamento por subcategoria do cartão   |
| `make run_budget_category`           | `ENVIRONMENT`, `DATE`, `CATEGORY` | Itens individuais de uma categoria                     |
| `make query_budget_unified`          | `ENVIRONMENT`, `DATE`             | Mesmo que `run_budget_unified` via sqlcmd direto no DB |
| `make query_balance`                 | `ENVIRONMENT`, `DATE`             | Mesmo que `run_balance` via sqlcmd direto no DB        |
| `make query_budget_cards_and_others` | `ENVIRONMENT`, `DATE`             | Mesmo que `run_budget_cards_and_others` via sqlcmd     |
| `make query_budget_category`         | `ENVIRONMENT`, `DATE`, `CATEGORY` | Mesmo que `run_budget_category` via sqlcmd             |
