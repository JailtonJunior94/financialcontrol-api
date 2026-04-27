Use a skill create-prd para definir a Modernização do Módulo de Cartões (Cards).

Entradas:
- Problema: O módulo `cards` atual possui uma estrutura que mistura camadas de infraestrutura e não implementa de forma robusta a regra de "melhor dia de compra". Além disso, a organização de pastas não está alinhada com os padrões de Clean Architecture desejados.
- Persona afetada: Desenvolvedores do sistema (manutenibilidade) e Usuários Finais (precisão financeira).
- Restricoes de escopo: Focado exclusivamente no módulo `internal/modules/cards`. Integração com `internal/modules/identity` apenas para validação de contexto de usuário via JWT. Listagem de bandeiras (flags) como recurso auxiliar.
- Restricoes tecnicas: Linguagem Go, Framework Web Fiber v2, Banco de Dados MSSQL. Princípios de DDD, Clean Architecture e Object Calisthenics (máximo 1 nível de indentação por método, sem 'else', wrap de primitivos).

Saidas esperadas obrigatorias:
- problema claro e verificavel: A estrutura atual do módulo `cards` dificulta a evolução e teste de regras de negócio complexas (como o cálculo de vencimento) devido ao acoplamento entre HTTP, Persistência e Domínio.
- objetivos mensuráveis:
    - 100% do código do módulo `cards` movido para a nova estrutura (`domain`, `application`, `infrastructure/http`, `infrastructure/persistence`).
    - 100% dos endpoints protegidos, garantindo que usuários vejam apenas seus próprios dados.
    - Zero violações de Object Calisthenics nas novas implementações.
- nao objetivos explicitos: Migração de banco de dados (permanece MSSQL) e alteração de outros módulos (exceto `identity` se necessário para infraestrutura comum).
- requisitos funcionais numerados (RF-01, RF-02...):
    - RF-01: CRUD de Cartões (Criar, Ler Lista/ID, Atualizar e Desativar).
    - RF-02: Listagem de Bandeiras disponíveis.
    - RF-03: Segurança por Usuário (extração de `user_id` do JWT e filtro obrigatório).
    - RF-04: Cálculo de Fatura (Lógica de "Melhor Dia de Compra" e "Data de Vencimento").
- requisitos nao funcionais (RNF-01, RNF-02...):
    - RNF-01: Clean Architecture (divisão em domain, application, infrastructure).
    - RNF-02: Object Calisthenics (sem 'else', 1 nível de indentação).
    - RNF-03: Injeção de Dependência (uso de interfaces para repositórios).
- criterios de aceite por requisito funcional:
    - RF-01/RF-03: O endpoint `GET /cards` retorna apenas cartões do usuário autenticado.
    - RF-04: Uma compra feita em 24/04 deve mostrar vencimento em 01/05.
    - RF-04: Uma compra feita em 25/04 deve mostrar vencimento em 01/06.
    - RNF-02: O código não deve possuir a palavra-chave `else`.
    - RNF-01: A estrutura de pastas deve seguir: `internal/modules/cards/infrastructure/persistence/mssql`.
- riscos com probabilidade e impacto:
    - Complexidade no cálculo de datas em meses com menos de 31 dias: Probabilidade Média, Impacto Alto.
    - Quebra de compatibilidade com o frontend atual: Probabilidade Baixa, Impacto Médio.
