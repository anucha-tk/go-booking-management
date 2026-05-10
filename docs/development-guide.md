# Development Guide

This guide provides instructions for developers to set up, run, and contribute to the Booking Management System.

## 🛠 Prerequisites
- **Go**: 1.26.2 or later
- **Docker & Docker Compose**: For running PostgreSQL
- **Task**: For running project commands (`brew install go-task/tap/go-task`)
- **SQLC**: For data access generation

## 🚀 Getting Started

1. **Clone the repository**
2. **Setup environment**:
   Copy `.env.example` to `.env` (if provided) or ensure `DB_URL` is configured in `Taskfile.yml`.
3. **Start Database**:
   ```bash
   task docker:up
   ```
4. **Run Migrations**:
   ```bash
   task db:migrate:up
   ```
5. **Seed Data** (Optional):
   ```bash
   task db:seed
   ```
6. **Run Application**:
   ```bash
   task run
   ```

## 🧪 Testing & Quality

### Running Tests
- **Unit Tests**: `task test`
- **Integration Tests**: `task itest` (requires Docker for Testcontainers)
- **Coverage**: `task test:coverage` (enforces **80%** threshold)

### Linting
We use `golangci-lint` to maintain code quality:
```bash
task lint
```

### Git Hooks
The project uses `lefthook` for pre-commit checks:
- Automatically runs `lint`, `test`, and `swagger:generate` before commit.

## 🔄 Development Workflow

1. **Database Changes**:
   - Add a new migration in `internal/infrastructure/db/migrations`.
   - Update SQL queries in `internal/infrastructure/db/queries`.
   - Run `task sqlc:generate`.
2. **API Changes**:
   - Update handlers and DTOs in `internal/adapter/http`.
   - Update Swagger annotations.
   - Run `task swagger:generate`.
3. **Logic Changes**:
   - Always follow TDD. Write tests in `*_test.go` files first.

## 📏 Coding Standards

- **Context First**: Always pass `context.Context` as the first argument for I/O operations.
- **Interface Segregation**: Depend on interfaces, not concrete implementations.
- **Naming**: Use clear, descriptive names. Avoid abbreviations.
- **Errors**: Wrap errors with context using `fmt.Errorf("...: %w", err)`.
