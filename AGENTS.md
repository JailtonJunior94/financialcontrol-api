# Regras para Agentes de IA

Este diretorio centraliza regras para uso com agentes de IA em tarefas reais de analise, alteracao e validacao de codigo.

## Objetivo

Use estas instrucoes para manter consistencia, seguranca e qualidade ao trabalhar com codigo, configuracao, validacao e evolucao de sistemas.

## Arquitetura: microservico

O projeto e um microservico independente de controle financeiro pessoal, com deploy isolado em Kubernetes via Docker. Expoe uma API REST (Fiber) e comandos CLI (Cobra) para geracao de orcamentos. Usa SQL Server (MSSQL) como banco de dados.

Stack detectada: Go 1.25.
Frameworks detectados: Fiber v2 (HTTP), Cobra (CLI), Viper (configuracao), sqlx + go-mssqldb (persistencia).
Infraestrutura: Docker, Kubernetes, GitHub Actions CI/CD, SonarCloud, Artillery (load testing).

## Estrutura de Pastas

```
.
├── main.go                              # Entrypoint: CLI (Cobra) com comandos server, budget, sync
├── go.mod / go.sum                      # Dependencias Go 1.25
├── Makefile                             # Atalhos: build, run, run_sync, run_budget, etc.
├── Dockerfile                           # Multi-stage build (golang → alpine)
├── docker-compose.yml                   # API + SQL Server local
├── config.{Development,Staging,Production}.yaml  # Configuracao por ambiente (Viper)
├── sonar-project.properties             # SonarCloud
├── cmd/
│   └── test-permissions/                # Utilitario auxiliar
├── src/
│   ├── domain/                          # Camada de dominio (sem dependencias externas)
│   │   ├── entities/                    # Entidades: bill, card, category, flag, invoice, transaction, user, budget
│   │   ├── interfaces/                  # Contratos de repositorio (IBillRepository, ICardRepository, etc.)
│   │   ├── usecases/                    # Contratos de servico (IBillService, IAuthService, etc.)
│   │   ├── events/                      # Event dispatcher + invoice_changed_event
│   │   ├── customErrors/                # Erros de dominio tipados
│   │   └── constants/                   # Constantes e rotas
│   ├── application/                     # Camada de aplicacao (orquestracao)
│   │   ├── services/                    # Implementacoes dos usecases (BillService, AuthService, etc.)
│   │   ├── usecase/                     # Casos de uso especificos (UpdateTransactionUseCase, etc.)
│   │   ├── handlers/                    # Listeners de eventos (InvoiceChangedListener)
│   │   ├── dtos/
│   │   │   ├── requests/                # DTOs de entrada
│   │   │   └── responses/               # DTOs de saida + HttpResponse padrao
│   │   └── mappings/                    # Mapeamento entidade ↔ DTO
│   ├── infrastructure/                  # Camada de infraestrutura
│   │   ├── repositories/                # Implementacoes dos contratos de repositorio
│   │   ├── adapters/                    # Adapters: hash (bcrypt), jwt, uuid
│   │   ├── database/                    # Conexao SQL (ISqlConnection)
│   │   ├── queries/                     # Queries SQL raw por entidade
│   │   ├── environments/                # Leitura de configuracao (Viper)
│   │   └── ioc/                         # Composicao de dependencias (manual DI)
│   ├── presentation/                    # Camada de apresentacao (HTTP)
│   │   ├── controllers/                 # Controllers Fiber por recurso
│   │   └── middlewares/                 # Middleware de autenticacao JWT
│   ├── app/                             # Bootstrap e setup do servidor
│   │   ├── server.go                    # RunServer (Fiber)
│   │   ├── budget.go                    # Comandos CLI de orcamento
│   │   ├── sync.go                      # Comando CLI de sincronizacao
│   │   ├── configuration/               # App factory + setup de rotas
│   │   └── routes/                      # Definicao de rotas por recurso
│   └── shared/                          # Utilitarios compartilhados (time, filter)
├── tests/
│   ├── artillery.yaml                   # Testes de carga
│   └── coverage.out                     # Cobertura de testes
├── .k8s/                                # Manifests Kubernetes
│   ├── deployments/                     # Deployment da API
│   ├── services/                        # Service (ClusterIP/NodePort)
│   ├── hpas/                            # HorizontalPodAutoscaler
│   ├── ingress/                         # Ingress
│   ├── certmanager/                     # Cert-Manager issuer (TLS)
│   └── namespaces/                      # Namespace financialcontrol
└── .github/
    ├── workflows/ci-cd.yml              # CI: test → SonarCloud → Docker build → K8s deploy
    ├── agents/                          # Wrappers de agentes (Copilot)
    └── skills/                          # Skills de governanca
```

## Padrao Arquitetural

Arquitetura em camadas (Layered Architecture) com influencias de Clean Architecture. As camadas sao bem definidas e o fluxo de dependencias aponta para dentro (dominio no centro).

### Camadas

| Camada | Pacote | Responsabilidade |
|---|---|---|
| **Domain** | `src/domain/` | Entidades, interfaces de repositorio, contratos de servico (usecases), eventos e erros de dominio. Sem dependencias externas. |
| **Application** | `src/application/` | Implementa contratos de servico, orquestra logica de negocio, define DTOs, mapeamentos e handlers de eventos. |
| **Infrastructure** | `src/infrastructure/` | Implementa contratos de repositorio e adapters. Contem queries SQL, conexao de banco, leitura de config e composicao de dependencias (IoC manual). |
| **Presentation** | `src/presentation/` | Controllers Fiber e middlewares HTTP. Depende apenas de contratos de servico do dominio. |
| **App/Bootstrap** | `src/app/` | Inicializacao do servidor, setup de rotas e comandos CLI. Ponto de composicao. |

### Fluxo de Dependencias

```
presentation/controllers → domain/usecases (interfaces de servico)
application/services     → domain/interfaces (interfaces de repositorio)
infrastructure/repos     → domain/interfaces (implementa contratos)
infrastructure/ioc       → todas as camadas (composicao raiz)
domain/                  → nenhuma dependencia externa
```

- Controllers dependem de interfaces de servico (`usecases`), nunca de implementacoes.
- Services dependem de interfaces de repositorio (`interfaces`), nunca de implementacoes.
- A camada de dominio e pura: sem imports de framework, banco ou HTTP.
- A composicao de dependencias ocorre em `infrastructure/ioc/dependency_injection.go` (manual, sem framework DI).

## Modo de trabalho

1. Entender o contexto antes de editar qualquer arquivo.
2. Preferir a menor mudanca segura que resolva a causa raiz.
3. Preservar arquitetura, convencoes e fronteiras ja existentes no contexto analisado.
4. Nao introduzir abstracoes, camadas ou dependencias sem demanda concreta.
5. Atualizar ou adicionar testes quando houver mudanca de comportamento.
6. Rodar validacoes proporcionais a mudanca.
7. Registrar bloqueios e suposicoes explicitamente quando o contexto estiver incompleto.

## Diretrizes de Estrutura

1. Priorize entendimento do codigo e do contexto atual antes de propor refatoracoes.
2. Respeite padroes existentes de nomenclatura, organizacao e tratamento de erro.
3. Defina estrutura simples, evolutiva e com defaults explicitos.
4. Evite reescritas amplas quando uma alteracao localizada resolver o problema.
5. Estabeleca contratos, testes e comandos de validacao cedo quando eles ainda nao existirem.
6. Considere risco de regressao como restricao principal.
7. Evite overengineering disfarcado de arquitetura futura.

## Regras por Arquitetura

1. Preservar contratos publicados e compatibilidade de integracao (rotas `/api/v1/*`).
2. Manter inicializacao, observabilidade e shutdown como parte do comportamento do servico.
3. Nao acoplar o servico a convencoes de outros servicos sem contrato explicito.
4. Respeitar a separacao de camadas: dominio nunca importa pacotes de infraestrutura, aplicacao ou apresentacao.
5. Novos repositorios devem definir contrato em `domain/interfaces/` e implementacao em `infrastructure/repositories/`.
6. Novos servicos devem definir contrato em `domain/usecases/` e implementacao em `application/services/`.
7. Registrar novas dependencias em `infrastructure/ioc/dependency_injection.go`.
8. Queries SQL ficam em `infrastructure/queries/`, nao espalhadas nos repositorios.
9. DTOs de request/response ficam em `application/dtos/`, nao no dominio.

## Regras por Linguagem

Para tarefas que alteram codigo, carregar a skill:

- `.agents/skills/agent-governance/SKILL.md`

Para tarefas que alteram codigo Go, carregar tambem:

- `.agents/skills/go-implementation/SKILL.md`

## Referencias da Skill

Ler conforme necessidade:

- `.agents/skills/agent-governance/references/ddd.md`
- `.agents/skills/agent-governance/references/error-handling.md`
- `.agents/skills/agent-governance/references/security.md`
- `.agents/skills/agent-governance/references/tests.md`

## Referencias da Skill Go

Ler conforme necessidade:

- `.agents/skills/go-implementation/references/governance.md`
- `.agents/skills/go-implementation/references/architecture.md`
- `.agents/skills/go-implementation/references/go-standards.md`
- `.agents/skills/go-implementation/references/interfaces.md`
- `.agents/skills/go-implementation/references/generics.md`
- `.agents/skills/go-implementation/references/concurrency.md`
- `.agents/skills/go-implementation/references/design-patterns.md`
- `.agents/skills/go-implementation/references/observability.md`
- `.agents/skills/go-implementation/references/api.md`
- `.agents/skills/go-implementation/references/persistence.md`
- `.agents/skills/go-implementation/references/configuration.md`
- `.agents/skills/go-implementation/references/implementation-examples.md`

## Validacao

Antes de concluir uma alteracao:

1. Rodar `gofmt` nos arquivos Go alterados.
2. Rodar primeiro testes direcionados e depois `go test ./...` quando o custo for proporcional.
3. Rodar `go vet ./...` quando esse passo fizer parte do gate do projeto.
4. Rodar lint se o contexto oferecer esse passo.
5. Informar falhas com o comando exato e um diagnostico curto.

## Restricoes

1. Nao inventar contexto ausente.
2. Nao assumir versao de linguagem, framework ou runtime sem verificar (Go 1.25, Fiber v2).
3. Nao alterar comportamento publico sem deixar isso explicito.
4. Nao usar exemplos como copia cega; adaptar ao contexto real.
5. Nao alterar contratos externos, readiness, observabilidade ou semantica operacional sem explicitar a mudanca.
6. Nao introduzir framework de DI; o projeto usa composicao manual em `ioc/`.
7. Nao mover queries SQL para dentro dos repositorios; manter em `infrastructure/queries/`.
