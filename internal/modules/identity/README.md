# identity

Módulo de identidade do Financial Control API. Implementa autenticação, gestão de usuários e propagação de identidade via `context.Context`.

## Propósito

Encapsula toda lógica relacionada à identidade: criação de conta, autenticação com emissão de JWT e consulta do usuário autenticado. Segue Clean Architecture / DDD com separação explícita entre domínio, aplicação e infraestrutura.

## Camadas

| Camada | Pacote | Responsabilidade |
|--------|--------|-----------------|
| Domínio | `domain/entities`, `domain/vos`, `domain/interfaces`, `domain/services`, `domain/factories` | Entidades, Value Objects, contratos de persistência/hash/token, regras puras |
| Aplicação | `application/usecase`, `application/dtos` | Orquestração de casos de uso via interfaces do domínio |
| Infraestrutura HTTP | `infrastructure/http/handlers`, `infrastructure/http/routes` | Handlers finos (parse/validate/serialize), registro de rotas |
| Infraestrutura Repositório | `infrastructure/repositories/mssql` | Persistência SQL Server via sqlx |
| Bootstrap | `module.go` | Construtor puro — `NewModule(Deps) *Module` — único ponto de wiring |

## Diagrama de camadas

```
HTTP Request
     │
     ▼
infrastructure/http/handlers   ← parse, validate, serialize, error mapping
     │
     ▼
application/usecase            ← orquestração, regras de aplicação
     │
     ▼
domain/{entities,services,vos} ← regras de domínio puras, sem IO
     │
     ▼
domain/interfaces              ← contratos: Hasher, TokenIssuer, UserRepository
     │
     ▼
infrastructure/repositories/mssql  ← SQL Server via sqlx
```

## Autenticação e JWT

- Emissão e validação de tokens usam exclusivamente `github.com/golang-jwt/jwt/v5` (via `pkg/jwt`).
- Algoritmo: HS256. Segredo: variável de ambiente `JWT_SECRET` (≥ 32 bytes). TTL: 15 min (configurável).
- Algoritmo `none` e qualquer alg divergente são rejeitados explicitamente.
- Claims: `sub` (UserID), `email`, `exp` — imutáveis (D-10/D-19).

## Propagação de identidade

O middleware `pkg/authmiddleware.Protected(parser)` extrai e valida o token e injeta `identitycontext.Identity` no `context.Context` via chave de tipo privado. Usecases e serviços leem a identidade via `pkg/identitycontext.FromContext(ctx)`, sem dependência de `*http.Request`.

## Bootstrap puro

O construtor `NewModule` recebe `Deps` por parâmetro e não usa auto-registro nem service locator:

```go
m := identity.NewModule(identity.Deps{
    DB:          sqlConn,
    Hasher:      identity.NewHasherAdapter(security.NewHashAdapter()),
    TokenIssuer: identity.NewTokenIssuerAdapter(jwtIssuer),
})
m.RegisterHTTP(router, authmiddleware.Protected(jwtParser))
```

## Value Objects

| VO | Invariante |
|----|-----------|
| `vos.Email` | Formato RFC 5322; rejeita vazio |
| `vos.UserID` | UUID v4; `New()` gera; `Parse(s)` valida formato |
| `vos.HashedPassword` | String não vazia; encapsula resultado de bcrypt |

## Endpoints

| Método | Path | Auth | Descrição |
|--------|------|------|-----------|
| `POST` | `/token` | — | Autentica e retorna `{ "token", "expires_at" }` |
| `GET` | `/me` | Bearer | Retorna dados do usuário autenticado lidos do `context.Context` |
| `POST` | `/users` | — | Cria usuário (idempotente por e-mail; 201 novo / 200 existente / 409 conflito) |

## Erros de domínio

Definidos em `domain/errors.go`; mapeados para HTTP em `infrastructure/http/handlers/error_mapping.go`. Nunca vazar detalhes internos na resposta. Falhas de auth logadas em nível `warn` sem dados sensíveis (D-17).

## Mocks

Gerados com `mockery v2` via `mockery.yml` na raiz. Regenerar com `make mocks`.

## Pontos de extensão

- Novo caso de uso: implementar interface em `application/usecase`, registrar em `module.go`.
- Novo provider de token: implementar `domain/interfaces.TokenIssuer`, adaptar em `module.go`.
- Novo repositório (ex.: PostgreSQL): criar em `infrastructure/repositories/postgres`, implementar `domain/interfaces.UserRepository`.
