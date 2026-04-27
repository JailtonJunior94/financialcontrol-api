Use a skill create-prd para definir a modernização do módulo de identity.

Entradas:
- Problema: O módulo `internal/modules/identity` apresenta acoplamento entre a camada de transporte (HTTP) e a lógica de negócio, estrutura de pastas inconsistente com Clean Architecture/DDD (ex: repositórios misturados na infraestrutura), uso de bibliotecas de JWT potencialmente desatualizadas e falta de propagação idiomática do token via `context.Context`.
- Persona afetada: Desenvolvedores Backend e Arquitetos de Software que buscam manutenibilidade, testabilidade e segurança no gerenciamento de identidade.
- Restricoes de escopo: A refatoração deve se limitar ao módulo `internal/modules/identity` e componentes compartilhados de segurança (JWT/Middleware). Não deve haver alterações em esquemas de banco de dados ou quebras de contrato nos endpoints existentes.
- Restricoes tecnicas: 
    - Implementação em Go; 
    - Aderência estrita a Clean Architecture e DDD seguindo o fluxo de camadas `handler -> usecase -> domain -> repository` (não utilizar controllers); 
    - Estrutura de pastas obrigatória dentro do módulo:
        - `application/dtos`
        - `application/usecase`
        - `domain/entities`, `domain/factories`, `domain/interfaces`, `domain/services`, `domain/vos`
        - `infrastructure/http`, `infrastructure/http/handlers`, `infrastructure/http/routes`
        - `infrastructure/repositories/<banco-usado>`
        - Raiz do módulo: `README.md` e `module.go` (para Injeção de Dependência)
    - Aplicação de princípios de Object Calisthenics (evitar else, níveis de indentação reduzidos, classes pequenas); 
    - Uso de bibliotecas modernas de JWT (ex: `golang-jwt/jwt/v5`); 
    - Obrigatoriedade de trafegar informações do token via `context.Context`.

Saidas esperadas obrigatorias:
- problema claro e verificavel
- objetivos mensuráveis
- nao objetivos explicitos
- requisitos funcionais numerados (RF-01, RF-02...)
- requisitos nao funcionais (RNF-01, RNF-02...)
- criterios de aceite por requisito funcional
- riscos com probabilidade e impacto
