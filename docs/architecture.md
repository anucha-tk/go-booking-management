# Technical Architecture

This document describes the architectural patterns and design decisions of the Go Booking Management System.

## 🏗 Architectural Style: Layered DDD

The project follows a **Domain-Driven Design (DDD)** inspired layered architecture. This ensures a clear separation of concerns and makes the system highly testable.

### 1. Domain Layer (`internal/domain`)
- **Entities**: Pure data structures (`User`, `Room`, `Booking`).
- **Interfaces**: Repository definitions that the application layer depends on.
- **Validation**: Business rules that don't depend on external systems.

### 2. Application Layer (`internal/application`)
- **Services**: Use cases that coordinate domain entities and repository interfaces.
- **Orchestration**: Logic like calculating booking prices and managing search providers.

### 3. Adapter Layer (`internal/adapter`)
- **HTTP**: Gin handlers, DTOs, and routing.
- **Persistence**: SQLC-generated code and repository implementations in `internal/infrastructure/db/sqlc`.

### 4. Infrastructure Layer (`internal/infrastructure`)
- Low-level concerns like database connection management, migrations, and external configuration.

## ⚡️ Key Technical Patterns

### Fan-out/Fan-in Search (NFR1)
The system supports multiple room search providers (currently defaulting to the database). The `RoomService.ListAvailableRooms` implements a concurrency pattern:
- **Fan-out**: Queries all registered providers in parallel.
- **Timeout**: Enforces a strict **200ms** timeout for the aggregate result.
- **Fan-in**: Aggregates unique results into a single list.

### Safe Concurrency & Overlap Check
Room availability is strictly enforced at the database level using SQL `OVERLAPS`:
```sql
WHERE NOT EXISTS (
    SELECT 1 FROM bookings
    WHERE room_id = $2
    AND status = 'confirmed'
    AND (
        (start_date, end_date) OVERLAPS ($3, $4)
    )
)
```
This prevents double-booking even under high concurrent load.

### Dependency Injection
Components are wired together manually in `cmd/api/main.go`. This makes it easy to swap implementations (e.g., using a mock repository in tests).

## 🛡 Security Architecture

- **Authentication**: JWT-based (RS256/HS256).
- **Password Hashing**: **Argon2id**, the winner of the Password Hashing Competition.
- **Token Revocation**: Implements a **JTI (JWT ID)** revocation list in PostgreSQL to allow immediate logout.
- **RBAC**: Role-Based Access Control enforced via Gin middleware (`RoleAdmin`, `RoleOfficer`, `RoleCustomer`).

## 📊 Non-Functional Requirements (NFRs)

| ID | Requirement | Implementation |
|----|-------------|----------------|
| NFR1 | Aggregator Timeout | 200ms context timeout in Room Service |
| NFR2 | Test Coverage | >80% threshold enforced in `Taskfile` |
| NFR3 | Performance | Optimized indexes for booking overlaps |
| NFR4 | Scalability | Stateless API, horizontally scalable |
