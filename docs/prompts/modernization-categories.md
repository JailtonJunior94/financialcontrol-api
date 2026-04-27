Use a skill create-prd para definir a Gestão de Categorias e Subcategorias.

Entradas:
- **Problema:** O gerenciamento de categorias atual é rudimentar, está acoplado ao módulo `catalog` e não suporta a hierarquia necessária para uma gestão financeira detalhada. É preciso modernizar o módulo para seguir padrões rigorosos de engenharia (DDD, Clean Arch, Object Calisthenics) e permitir a organização em categorias e subcategorias, garantindo que o código seja testável e escalável.
- **Persona afetada:** Desenvolvedores do backend (manutenibilidade) e usuários finais do sistema financeiro (organização de gastos).
- **Restrições de escopo:** 
    - Criação do novo módulo em `internal/modules/categories`.
    - Implementação de categorias principais e subcategorias vinculadas.
    - CRUD completo com Soft Delete.
    - Listagem paginada com filtro por nome.
    - Exclui-se: Migração automática de dados legados e interface de usuário (frontend).
- **Restrições técnicas:** 
    - Linguagem: Go (Golang).
    - Persistência: SQL Server.
    - Arquitetura: Clean Architecture com separação em `domain`, `application` e `infrastructure`.
    - Regras de Qualidade: Object Calisthenics (máximo 1 nível de indentação por método, sem else, classes pequenas).
    - Padrões: Domain Services para regras de negócio complexas e Design Patterns (Repository, Factory, etc).
    - Localização: Handlers HTTP em `infrastructure/http` e Repositórios em `infrastructure/repositories`.

Saidas esperadas obrigatorias:

1. **Problema claro e verificável:**
    - Descrição do estado atual (acoplamento e falta de hierarquia) vs estado desejado (módulo isolado com subcategorias).
    - Justificativa técnica para o uso de Soft Delete e Domain Services.

2. **Objetivos mensuráveis:**
    - [ ] Módulo `internal/modules/categories` criado com estrutura de subpastas completa.
    - [ ] Endpoints de CRUD de Categorias (Pai) e Subcategorias (Filho) operacionais.
    - [ ] Filtro de busca por nome retornando resultados paginados.
    - [ ] Implementação de Soft Delete validada (registros não aparecem em queries comuns mas persistem no banco).

3. **Nao objetivos explicitos:**
    - Não implementar autenticação (utilizar middleware existente).
    - Não realizar refatoração em outros módulos além do `catalog` (apenas para desacoplamento).
    - Não criar relatórios ou gráficos nesta fase.

4. **Requisitos funcionais numerados:**
    - **RF-01:** O sistema deve permitir o cadastro de categorias principais (ex: Custos Fixos, Conforto).
    - **RF-02:** O sistema deve permitir o cadastro de subcategorias vinculadas a uma categoria principal (ex: Supermercado vinculado a Custos Fixos).
    - **RF-03:** O sistema deve permitir a edição de nomes e ícones/cores de categorias e subcategorias.
    - **RF-04:** O sistema deve implementar Soft Delete em categorias: ao "excluir" uma categoria pai, deve-se decidir (via regra de negócio) se as subcategorias são marcadas como excluídas ou se a operação é impedida.
    - **RF-05:** O sistema deve listar categorias e subcategorias de forma hierárquica ou plana com filtros e paginação.
    - **RF-06:** O sistema deve validar que uma subcategoria sempre possua um vínculo com uma categoria pai ativa.

5. **Requisitos nao funcionais (RNF):**
    - **RNF-01:** O tempo de resposta da listagem paginada não deve exceder 200ms para bases de até 10k registros.
    - **RNF-02:** O código deve atingir no mínimo 80% de cobertura de testes unitários no `domain` e `application`.
    - **RNF-03:** A estrutura deve seguir estritamente o Object Calisthenics para garantir baixo acoplamento.

6. **Criterios de aceite por requisito funcional:**
    - **CA (RF-01/02):** Payload de criação validado (campos obrigatórios presentes).
    - **CA (RF-04):** Query SQL gerada deve conter `WHERE deleted_at IS NULL`.
    - **CA (RF-05):** Parâmetros `page` e `pageSize` obrigatórios na listagem, com defaults (1 e 10).

7. **Riscos com probabilidade e impacto:**
    - **Risco 01:** Complexidade na deleção em cascata (Soft Delete). 
        - Probabilidade: Média | Impacto: Alto.
        - Mitigação: Uso de Domain Services para orquestrar a deleção segura.
    - **Risco 02:** Performance em queries recursivas ou de muitos níveis (caso a hierarquia cresça).
        - Probabilidade: Baixa | Impacto: Médio.
        - Mitigação: Limitar a hierarquia a 2 níveis (Pai/Filho) conforme solicitado.
