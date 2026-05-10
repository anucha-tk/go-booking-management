# Go Booking Management System

A robust, enterprise-grade booking management system built with Go, focusing on high performance, maintainability, and architectural excellence.

## 🚀 Overview

This system provides a comprehensive set of APIs for managing room bookings, featuring:

- **JWT-based Authentication** with token revocation (JTI).
- **Role-Based Access Control (RBAC)** for users and administrative officers.
- **Room Management** with detailed views and booking history.
- **Concurrency Control** to prevent double bookings.
- **Structured Logging** and **Rate Limiting**.
- **Automated API Documentation** with Swagger/Scalar.

## 🛠 Tech Stack

- **Core**: Go (Golang) 1.21+
- **Web Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **ORM/Query Generator**: [SQLC](https://sqlc.dev/) (Type-safe SQL)
- **Migrations**: [Golang Migrate](https://github.com/golang-migrate/migrate)
- **Authentication**: JWT (JSON Web Tokens)
- **Documentation**: [Swagger](https://github.com/swaggo/swag) / [Scalar](https://github.com/scalar/scalar)
- **Task Runner**: [Taskfile](https://taskfile.dev/)
- **Live Reload**: [Air](https://github.com/cosmtrek/air)
- **Linting**: [GolangCI-Lint](https://golangci-lint.run/)

## 📖 Documentation

Detailed documentation is available in the `docs/` directory:

- [**Architecture**](docs/architecture.md): System design, patterns, and architectural decisions.
- [**API Guide**](docs/api-guide.md): Comprehensive guide to using the available APIs.
- [**Data Models**](docs/data-models.md): Database schema and entity relationships.
- [**Development Guide**](docs/development-guide.md): Setup instructions and workflow patterns.
- [**Project Index**](docs/index.md): Entry point for all project documentation.

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- [Task](https://taskfile.dev/installation/) (Task runner)

### Setup

1. **Clone the repository**
2. **Setup environment variables**
   ```bash
   cp .env.example .env
   ```
3. **Start the database**
   ```bash
   task docker:up
   ```
4. **Run migrations**
   ```bash
   task db:migrate:up
   ```
5. **Seed the database (Optional)**
   ```bash
   task db:seed
   ```
6. **Run the application**
   ```bash
   task run
   ```

## 🛠 Development Tasks

We use `Taskfile` for common development operations:

| Task                      | Description                        |
| :------------------------ | :--------------------------------- |
| `task run`                | Run the API server                 |
| `task watch`              | Run with live reload (Air)         |
| `task build`              | Build the binary                   |
| `task test`               | Run all unit tests                 |
| `task test:coverage`      | Run tests and verify 80% threshold |
| `task lint`               | Run golangci-lint                  |
| `task swagger:generate`   | Update API documentation           |
| `task docker:up` / `down` | Manage DB container                |

## 🧪 Testing

The project maintains a high quality standard with **>80% test coverage**.

```bash
# Run tests and ensure coverage threshold
task test:coverage

# Run integration tests
task itest
```

## 🔒 Security

- **Rate Limiting**: Protected against brute force and DDoS.
- **JWT Revocation**: Secure logout functionality using JTI tracking.
- **Input Validation**: Strict DTO validation for all API requests.
- **Role Isolation**: Sensitive operations restricted to `RoleOfficer`.
