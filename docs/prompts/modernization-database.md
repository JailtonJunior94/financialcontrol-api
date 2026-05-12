# Modernização de Banco de Dados e Migrações com devkit-go

Use a skill [create-prd] para definir a integração do [devkit-go] e comando de [migration].

Entradas:
- Problema: A camada de persistência atual utiliza uma implementação customizada baseada em `sqlx` que não segue os padrões do ecossistema `devkit-go`. Além disso, a execução de migrações não está isolada do ciclo de vida da aplicação principal, dificultando o deploy em ambientes orquestrados (K8s/Docker) onde a migration deve rodar como um Init Container ou Job prévio.
- Persona afetada: Desenvolvedores Backend e Engenheiros de SRE/DevOps.
- Restricoes de escopo: 
    - Implementação restrita a MSSQL.
    - Criação de CLI exclusivamente para execução de migrações.
    - Não inclui a refatoração de todos os repositórios da aplicação (apenas a infraestrutura de conexão).
- Restricoes tecnicas: 
    - Linguagem: Go 1.26.2.
    - Dependência: `github.com/JailtonJunior94/devkit-go` v0.4.0.
    - O comando CLI deve ser implementado em `cmd/migration`.
    - Deve suportar as variáveis de ambiente já existentes para configuração de DB.

Saidas esperadas obrigatorias:
- problema claro e verificavel: Acoplamento da lógica de migração e falta de padronização da camada de DB.
- objetivos mensuráveis: 
    - 100% das migrações executadas via CLI.
    - Redução de boilerplate na configuração de conexão MSSQL usando `devkit-go`.
- nao objetivos explicitos: 
    - Não refatorar a lógica de negócio ou modelos de entidade.
    - Não adicionar suporte a outros bancos (Postgres/MySQL) neste momento.
- requisitos funcionais numerados:
    - RF-01: Configurar o provedor de MSSQL utilizando `devkit-go/pkg/database`.
    - RF-02: Implementar o comando CLI em `cmd/migration/main.go` para processar arquivos `.sql` da pasta `/migrations`.
    - RF-03: Garantir que o comando CLI retorne exit code 0 em sucesso e != 0 em falha para controle de pipeline.
    - RF-04: Integrar o mecanismo de Migrations do `devkit-go`.
- requisitos nao funcionais:
    - RNF-01: O comando de migration deve ser idempotente.
    - RNF-02: O tempo de inicialização da conexão não deve impactar a performance do CLI.
    - RNF-03: Logs claros indicando quais migrações foram aplicadas ou se já estavam atualizadas.
- criterios de aceite por requisito funcional:
    - RF-01: A aplicação deve conectar ao MSSQL usando a abstração do `devkit-go`.
    - RF-02: Executar `./migration` deve aplicar os scripts da pasta `migrations` no banco de dados.
    - RF-03: Em caso de erro de sintaxe no SQL, o CLI deve parar a execução e retornar erro.
    - RF-04: Deve ser possível rodar as migrations independentemente do `main.go` da API.
- riscos com probabilidade e impacto:
    - Risco: Incompatibilidade de versões do driver MSSQL (Probabilidade: Baixa / Impacto: Alto).
    - Risco: Travamento de tabelas durante migrações em produção (Probabilidade: Média / Impacto: Muito Alto).
