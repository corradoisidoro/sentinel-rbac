# Sentinel RBAC 🔐

[![Go](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)]()

Sentinel RBAC is a clean, production-grade REST API built in Go that demonstrates best practices for authentication, authorization, rate limiting, and service architecture.
A Go backend service featuring secure JWT authentication and granular Role-Based Access Control (RBAC). Designed with security-first middleware to protect sensitive administrative endpoints.

✨ Key Highlights
- 🔑 JWT Authentication
- 🛂 Role-Based Access Control (RBAC)
- 🚦 Multi-Layer Rate Limiting (Global, IP, Route)
- 🧱 Clean Architecture (Handler → Service → Repository)
- 🛡️ Security-Focused Design
- 🔄 Graceful Shutdown
- 🧪 Testable & Deterministic Middleware
- 🗄️ Database Migrations with GORM
- ⚙️ Config-Driven Setup

🧠 Why This Project Exists
This project was built to demonstrate:
- How I design maintainable Go services
- How I think about security and abuse prevention
- How I balance simplicity vs production readiness
- How I structure APIs that scale beyond MVPs

It avoids unnecessary frameworks and over-engineering while still addressing real production concerns.

🏗️ Architecture Overview
```
cmd/
└── main.go              # Application entrypoint

internal/
├── config/              # Configuration loading & validation
├── handler/             # HTTP handlers (Gin)
├── middleware/          # Auth, RBAC, Rate Limiting
├── models/              # Database models
├── repository/          # Data access layer
└── service/             # Business logic
```

🚦 Rate Limiting Strategy
Sentinel RBAC implements multi-layer rate limiting using golang.org/x/time/rate:
| Layer     | Purpose                      |
| --------- | ---------------------------- |
| Global    | Protects server capacity     |
| Per-IP    | Prevents abuse               |
| Per-Route | Protects expensive endpoints |

🚀 Running the Project

Prerequisites
- Go 1.21+
- Git

Clone & Run
```bash
git clone https://github.com/corradoisidoro/sentinel-rbac.git
cd sentinel-rbac
go run ./cmd
```

Environment Variables
```
DATABASE_URL=sentinel.db
JWT_SECRET=super-secret-key
SERVER_PORT=8080
```

📡 API Endpoints

Public
- GET /ping — Health check
- POST /api/auth/register
- POST /api/auth/login

Authenticated
- POST /api/auth/logout
- GET /api/users/profile

Admin Only
- GET /api/users/admin

🧪 Testing
```bash
go test ./...
go test ./... -v
go test -race ./...
```

🧰 Tech Stack

- Language: Go
- Framework: Gin
- ORM: GORM
- Auth: JWT
- Rate Limiting: golang.org/x/time/rate
- Database: SQLite (portable)