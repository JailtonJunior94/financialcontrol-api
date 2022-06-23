# Financial Control API

[![Go](https://github.com/JailtonJunior94/financialcontrol-api/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/JailtonJunior94/financialcontrol-api/actions/workflows/ci-cd.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=alert_status)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=bugs)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=code_smells)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=JailtonJunior94_financialcontrol-api&metric=coverage)](https://sonarcloud.io/dashboard?id=JailtonJunior94_financialcontrol-api)

## Sobre
Projeto backend de controle de finanças pessoais.

## Tecnologias Utilizadas 🚀
* **[Golang](https://golang.org/)**
* **[Heroku](https://dashboard.heroku.com/)**
* **[GitHub Actions](https://docs.github.com/pt/actions)**
* **[Docker](https://www.docker.com/)**
* **[SQL Server](https://www.microsoft.com/pt-br/sql-server/sql-server-2019)**

## Testes de Unidade
Para gerar o arquivo coverage da aplicação
```
go test --coverprofile tests/coverage.txt ./...
go test --coverprofile tests/coverage.out ./...
```
Para gerar html com informações detalhadas do teste
```
go tool cover --html=tests/coverage.txt
go tool cover --html=tests/coverage.out
```

docker image build -t jailtonjunior/financialcontrol:v1 .

docker image push jailtonjunior/financialcontrol:v1

kubectl get certificate -n financialcontrol
kubectl describe certificate -n financialcontrol
kubectl get certificaterequest -n financialcontrol

### Atualização de tabela 
```
ALTER TABLE dbo.Invoice 
ADD MarkImportTransactions BIT NULL
DEFAULT 0

UPDATE dbo.Invoice SET MarkImportTransactions = 0
```