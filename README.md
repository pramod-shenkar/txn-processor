# Internal Transfers Service (Golang)

A lightweight, concurrency-safe internal transfers system built using Go, MariaDB, and Redis.  
Implements account creation, balance queries, and atomic internal transfers.

---

## 🔧 Architecture & Key Decisions

### ✔ Hexagonal Architecture (Ports & Adapters)
Clear separation between:
- HTTP handlers (Fiber)
- Service layer
- DAO layer (GORM + Redis)
- Ports (inbound/outbound interfaces)
- Domain models

### ✔ Distributed Tracing (OpenTelemetry + Jaeger)
Tracing added at:
- HTTP middleware
- Service layer
- DAO operations

### ✔ ACID-Safe Transfers (SELECT FOR UPDATE)
Transfers run in a single DB transaction:
- Row-level locks on both accounts  
- Balance validation  
- Atomic update  
Prevents race conditions and double spending.

### ✔ Concurrency Optimizations
- **errgroup** to load both accounts in parallel  
- **errgroup** to update all caches concurrently after commit  
- **singleflight** to avoid cache stampede on account reads

### ✔ Redis Cache-aside Strategy
- DB is the source of truth  
- Cache updated after commit  
- On cache write error → key is deleted to avoid stale data  

### ✔ E2E Tests Using Testcontainers
- MariaDB container  
- Redis container  
- Full flow tested: create → transfer → verify balances  

### ✔ Load Testing with hey
Used to stress-test transfer endpoint under concurrency.

---

## 📦 Requirements

- Go 1.22+
- Docker (for MariaDB, Redis, Jaeger)
- Taskfile (`go install github.com/go-task/task/v3/cmd/task@latest`)

---

## 🚀 Running the Application

Start DB, Redis, Jaeger (optional), then run:

```bash
task run
```
 Running Load
```bash
task load
```

## 📘 API Endpoints

Create Account
```bash
curl -X POST http://localhost:9999/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id":1001,"initial_balance":"500"}'
```
Get Account
```bash
curl http://localhost:9999/v1/accounts/1001
```
Transfer
```bash
curl -X POST http://localhost:9999/v1/transfers \
  -H "Content-Type: application/json" \
  -d '{"source_account_id":1001,"destination_account_id":2002,"amount":"150"}'
```

