# AI Agent Context — primeTradingBackend

> **Last updated:** 2026-03-19  
> This file is the source of truth for any AI agent resuming work on this project.
> Keep it updated after every significant change.

---

## Project Purpose

A Go REST API backend for a trading/finance application. It provides:
- User authentication (register, login, JWT via HttpOnly cookie)
- Live commodity price fetching (gold/silver) from the goldpricez.com API
- Correlation data management between commodities (Pearson R / Spearman Rho)

---

## Architecture: Hexagonal (Ports & Adapters)

```
cmd/server/main.go            ← Wiring / Dependency Injection / Router setup
internal/
  domain/
    model/                    ← Pure Go structs (User, Commodity, Correlation)
    repository/               ← Outbound port interfaces (UserRepository, CorrelationRepository, CommodityRepository)
    algorithm/                ← (empty) Planned location for correlation algorithms
  application/                ← Business logic / Use-case services
    user_service.go           ← UserService struct + LoginInput / RegisterInput types
    login_service.go          ← Login() and RefreshToken() methods on UserService
    register_service.go       ← Register() method on UserService
    commodity_service.go      ← CommodityService; defines MetalPriceProvider port interface
  adapters/
    postgres/                 ← Implementations of domain repository interfaces
      postgres.go             ← NewPostgresDB() connection helper
      user_repo.go            ← Implements repository.UserRepository (+ Migrate)
      correlation_repo.go     ← Implements repository.CorrelationRepository (+ Migrate)
      commodity_repo.go       ← Implements repository.CommodityRepository (stub/partial)
    goldpricez/               ← External API adapter
      client.go               ← Implements MetalPriceProvider; all HTTP logic lives here
  handler/                    ← HTTP handlers (inbound adapters)
    helpers.go                ← jsonError() helper — all errors return consistent JSON
    user_handler.go           ← UserHandler; depends on UserServicePort interface
    stock_handler.go          ← CommodityHandler; depends on CommodityServicePort interface
    dto/                      ← Request/Response structs (LoginRequest, LoginResponse, etc.)
  auth/                       ← JWT generation and verification
  middleware/                 ← JWTAuthMiddleware, AdminRoleMiddleware
  errors/                     ← Custom ValidationError types
  client/                     ← (empty) Not yet used
utils/                        ← Validator helpers (username, email, password rules)
tests/                        ← Integration/handler tests
```

### Dependency Rule
**Domain → nothing. Application → Domain. Adapters → Domain/Application. Handlers → Application (via interfaces).**  
No layer may import a layer above it.

---

## Key Design Decisions

| Decision | Detail |
|---|---|
| **Service interfaces in handler package** | `UserServicePort` and `CommodityServicePort` are defined in the handler package so handlers aren't coupled to concrete types |
| **`MetalPriceProvider` port** | Defined in `application/commodity_service.go`; implemented by `adapters/goldpricez/Client` |
| **Migrations on repos** | `Migrate()` is on the concrete postgres struct. It is in `repository.UserRepository` interface (by design) but NOT in `CorrelationRepository` interface — called via concrete type in `main.go` when needed |
| **JWT via HttpOnly cookie** | Token set on login, cleared on logout, refreshable via `/api/refresh` |
| **No ORM** | Raw `database/sql` with `*sql.DB` passed into repos |

---

## What Is Done ✅

### Authentication
- [x] `POST /api/register` — validates, hashes password with bcrypt, saves user
- [x] `POST /api/login` — validates credentials, returns JWT in HttpOnly cookie
- [x] `POST /api/logout` — clears the cookie
- [x] `POST /api/refresh` — verifies existing token, issues a new one
- [x] `GET /api/me` — protected endpoint; returns current user info from JWT context
- [x] `JWTAuthMiddleware` — reads cookie or Bearer header, injects claims into context
- [x] `AdminRoleMiddleware` — skeleton in place, routes group wired

### Commodity Prices
- [x] `GET /api/commodity?type=gold|silver` — fetches live price from goldpricez.com API
- [x] Background ticker (every 3 min) calls `UpdateMetalPrices()` to refresh both metals
- [x] `CommodityService` is pure logic; HTTP call lives in `adapters/goldpricez/client.go`

### Correlation Repository
- [x] `correlations` table DDL defined (in `Migrate()`)
- [x] `Save(correlation)` — INSERT one record
- [x] `SaveBatch(correlations)` — INSERT multiple records in a loop
- [x] `GetLatest(a, b)` — most recent correlation between two commodities
- [x] `GetHistory(a, b, limit)` — paginated history, DESC by date
- [x] `GetTopCorrelated(commodity, limit)` — strongest pairs by ABS(pearsonR)

### Infrastructure
- [x] PostgreSQL connection via `postgres.NewPostgresDB()`
- [x] Docker + `docker-compose.yml` for local DB
- [x] Air (`.air.toml`) for hot-reload in development
- [x] CORS configured for `localhost:3100` and the Vercel frontend
- [x] Consistent JSON error responses via `jsonError()` helper

---

## What Is NOT Done / In Progress ⏳

### Correlation Feature (main next area of work)
- [ ] **Correlation calculation algorithm** — `internal/domain/algorithm/` is empty. Need to implement Pearson R and Spearman Rho calculation logic
- [ ] **Correlation service** — No `CorrelationService` in `application/` yet; needs to orchestrate fetching prices, running the algorithm, and persisting results
- [ ] **Correlation handler** — No HTTP endpoint for correlations yet
- [ ] **`CommodityRepository`** — Interface exists, postgres adapter exists (`commodity_repo.go`) but implementation methods appear to be stubs; needs full implementation before the correlation service can use stored price data

### Commodity Persistence
- [ ] `UpdateMetalPrices()` fetches but **does not persist** prices to the DB yet
- [ ] `commodity_repo.go` needs full `Save` / `GetLatestPrice` / `GetPriceHistory` implementation
- [ ] The `commodities` table migration has not been wired into `main.go`

### Known Bugs / Issues (from WHATSNEXT.md)
- [ ] Registration: PostgreSQL error `duplicate key value violates unique constraint "user_pkey"` is being surfaced raw to the frontend — should be caught and returned as a friendly `409 Conflict`
- [ ] Login: Frontend receives non-JSON response in some error paths, causing `JSON.parse` failure — likely a code path that still uses `http.Error` (plain text) rather than `jsonError`

### Other
- [ ] `internal/client/` is empty — may be intended as a future home for other external API clients
- [ ] No integration tests for commodity or correlation flows
- [ ] `Secure: false` on JWT cookie — must be set to `true` before production deployment

---

## Environment Variables Required

| Variable | Purpose |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | PostgreSQL connection |
| `JWT_SIGNING_KEY` | JWT signing secret |
| `API_KEY` | goldpricez.com API key |
| `HOST`, `PORT` | Server bind address (default `:8080`) |

---

## Running Locally

```bash
# Start the database
docker-compose up -d

# Run with hot reload
air

# Or run directly
go run ./cmd/server/main.go
```
