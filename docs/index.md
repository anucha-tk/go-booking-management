# Booking Management System Documentation

Welcome to the technical documentation for the **Go Booking Management System**. This project is a robust, production-ready backend system for managing room bookings, featuring a clean architecture, secure authentication, and high-performance room searching.

## 🚀 Quick Start

- **Technology Stack**: Go 1.26.2, Gin, PostgreSQL (sqlc)
- **API Documentation**: Available at `/v1/doc` (via Scalar) when running locally.
- **Port**: 8088

## 📚 Knowledge Base

### [Architecture & Design](./architecture.md)
Learn about the DDD-inspired layered architecture, dependency injection patterns, and the Fan-out/Fan-in search mechanism.

### [Data Models & Schema](./data-models.md)
Detailed breakdown of the core entities (`User`, `Room`, `Booking`) and their relationships in the PostgreSQL database.

### [API Guide](./api-guide.md)
Comprehensive list of available endpoints, authentication flows, and security measures.

### [Development Guide](./development-guide.md)
Instructions on how to set up the development environment, run tests (unit & integration), and follow the coding standards.

## 🛠 Project Structure

```text
.
├── api/                # Generated OpenAPI/Swagger specifications
├── cmd/                # Entry points (api server, seeding tools)
├── docs/               # Technical documentation (this folder)
├── internal/           # Private application code
│   ├── adapter/        # Implementation of external interfaces (HTTP, DB)
│   ├── application/    # Orchestration of business logic (Use Cases)
│   ├── domain/         # Core business logic and entity definitions
│   ├── infrastructure/ # Low-level tools (DB migrations, SQLC)
│   └── server/         # Server initialization logic
├── pkg/                # Public shared packages (logger, auth)
└── Taskfile.yml        # Task runner for build, test, and dev commands
```

## 🎯 Project Goals & Standards

- **Clean Code**: Strict adherence to SOLID principles and DDD.
- **Test Coverage**: Minimum **80%** threshold enforced via CI.
- **Security**: Argon2id password hashing and JTI-based JWT revocation.
- **Performance**: 200ms aggregator timeout for room search.

---
*Last Updated: 2026-05-10*
