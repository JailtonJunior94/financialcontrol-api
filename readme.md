# Financial Control API

[![Go](https://github.com/JailtonJunior94/financialcontrol-api/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/JailtonJunior94/financialcontrol-api/actions/workflows/ci-cd.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=alert_status)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=bugs)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=code_smells)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=coverage)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)

API backend para controle de financas pessoais, organizada em um layout Go-like com entrypoint unico em `cmd/financialcontrol-api`.

## Visao geral da estrutura

```text
.
|-- cmd/
|   |-- financialcontrol-api/   # entrypoint do binario principal
|   `-- test-permissions/       # binario auxiliar de teste
|-- configs/                    # config.<ENV>.yaml
|-- deployments/
|   |-- docker/                 # Dockerfile e compose
|   `-- k8s/                    # manifests Kubernetes
|-- internal/
|   |-- application/            # DTOs, services, handlers e use cases
|   |-- bootstrap/              # wiring de CLI, HTTP e container
|   |-- domain/                 # entidades, eventos e contratos
|   |-- http/                   # middlewares e constantes HTTP compartilhadas
|   |-- infrastructure/         # config, banco, queries e repositories
|   `-- shared/                 # utilitarios internos do servico
|-- tests/                      # artefatos auxiliares de teste e coverage
|-- Makefile
`-- .github/workflows/ci-cd.yml
```

## Entry point e fluxo de execucao

- Binario principal: `cmd/financialcontrol-api/main.go`
- Bootstrap de CLI: `internal/bootstrap/cli`
- Bootstrap HTTP: `internal/bootstrap/http`
- Composicao de dependencias: `internal/bootstrap/container`

O binario sobe a API por padrao e tambem expoe os comandos operacionais `sync`, `budget`, `budget-cards-and-others`, `budget-unified`, `budget-full`, `balance` e `budget-category`.

Para listar os comandos disponiveis:

```bash
go run ./cmd/financialcontrol-api --help
```

## Configuracao

Os arquivos de configuracao ficam em `configs/`:

- `configs/config.Development.yaml`
- `configs/config.Staging.yaml`
- `configs/config.Production.yaml`

O runtime exige a variavel `ENVIRONMENT` com um dos nomes acima. O loader resolve os arquivos a partir de `configs/` e falha cedo quando o arquivo nao existe.

Exemplos:

```bash
ENVIRONMENT=Development go run ./cmd/financialcontrol-api
ENVIRONMENT=Production go run ./cmd/financialcontrol-api sync
```

## Desenvolvimento local

Build do binario:

```bash
make build
```

Subir a API localmente:

```bash
make run ENVIRONMENT=Development
```

Executar sync:

```bash
make run_sync ENVIRONMENT=Production
```

Executar comandos de orcamento:

```bash
make run_budget ENVIRONMENT=Production DATE=01/04/2025
make run_budget_cards_and_others ENVIRONMENT=Production DATE=01/04/2025
make run_budget_unified ENVIRONMENT=Production DATE=01/04/2025
make run_budget_full ENVIRONMENT=Production DATE=01/04/2025
make run_balance ENVIRONMENT=Production DATE=01/04/2025
make run_budget_category ENVIRONMENT=Production DATE=01/04/2025 CATEGORY=Alimentacao
```

## Validacao

Rodar a suite de testes:

```bash
make test
```

Rodar validacao estatica disponivel no projeto:

```bash
make vet
```

Gerar HTML de cobertura com base no artefato produzido pelo `make test`:

```bash
go tool cover --html=tests/coverage.out
```

## Docker e operacao

Os artefatos operacionais agora ficam em `deployments/`:

- Dockerfile: `deployments/docker/Dockerfile`
- Compose: `deployments/docker/docker-compose.yml`
- Kubernetes: `deployments/k8s/`

Build da imagem local:

```bash
docker build -f deployments/docker/Dockerfile -t financialcontrol-api:local .
```

Subir stack local com Docker Compose:

```bash
docker compose -f deployments/docker/docker-compose.yml up --build
```

## CI/CD

O workflow em `.github/workflows/ci-cd.yml` executa:

1. `make build`
2. `make test`
3. SonarCloud scan
4. substituicao de variaveis em `configs/config.Production.yaml`
5. build e push da imagem a partir de `deployments/docker/Dockerfile`
6. deploy dos manifests em `deployments/k8s/`

## Contrato HTTP

O contrato publico permanece sob o prefixo `/api/v1`. Os testes de bootstrap e de registro no composition point ativo em `internal/bootstrap/http/http_test.go` e `internal/platform/http/router_test.go` ajudam a garantir a preservacao dos endpoints durante a reorganizacao estrutural.
