# Indico Technical Test – Backend

This repository contains the **backend API** for the Indico Technical Test.  
Built with **Go (Golang)** using the **Gin Framework**, PostgreSQL, and **SQLC** for safe, type-safe database access.

It provides a complete REST API for managing vouchers, including authentication, CSV import/export, and pagination + sorting.

---

## 🌐 Live Backend

Production API runs behind Docker on VPS:

[https://techtest-indico-be.alifdwt.com](https://techtest-indico-be.alifdwt.com)

Swagger UI:

[https://techtest-indico-be.alifdwt.com/swagger/index.html](https://techtest-indico-be.alifdwt.com/swagger/index.html)

---

## 📌 Features

### 1. Authentication

- Login with email & password
- Returns JWT token
- Secured endpoints using Bearer Token
- Gin middleware for route protection

### 2. Voucher Management

- Create voucher
- Update voucher
- Delete voucher
- Get voucher by ID
- List vouchers with:
  - Search by voucher code
  - Pagination
  - Sorting by:
    - `expiry_date`
    - `discount_percent`
    - `created_at`
    - `updated_at`

### 3. CSV Upload

- Upload bulk vouchers from CSV
- Header order is flexible
- Returns detailed failure reports per row:
  - Row number
  - Voucher code
  - Reason for failure

### 4. CSV Export

- Export all vouchers to CSV
- Format:
  ```csv
  voucher_code,discount_percent,expiry_date
  ```

---

## 🏗 Tech Stack

- Go (Golang)
- Gin Framework
- PostgreSQL
- SQLC (query → type-safe Go code)
- Docker & Docker Compose
- Swagger (OpenAPI 2.0)

---

## 🗂 Project Structure

```
.
├── cmd/api            # App entry point
│   └── main.go
├── docs               # Swagger docs
│   ├── swagger.yaml
│   └── swagger.json
├── internal
│   ├── config         # Config & logger
│   ├── dto            # Data Transfer Objects + validation
│   ├── handler        # HTTP handlers (controllers)
│   ├── middleware     # Auth middleware
│   ├── repository     # Database access (SQLC generated)
│   ├── routes         # Route registration
│   ├── service        # Business logic layer
│   └── util           # Shared helpers (response wrapper)
├── db
│   ├── migration      # SQL migrations
│   └── query          # SQL queries for SQLC
├── scripts            # Deployment helper scripts
├── filetest           # Sample CSV files
├── Dockerfile
├── docker-compose.yml
├── go.mod / go.sum
└── Makefile
```

---

## 🧠 Architecture Overview

This backend follows a clean-ish layered architecture:

```
HTTP Request
   ↓
Routes → Middleware
   ↓
Handlers
   ↓
Services (business logic)
   ↓
Repositories (SQLC)
   ↓
PostgreSQL
```

Benefits:

- Separation of concerns
- Easier testing
- Easy to extend for new features

---

## ⚙️ Environment Variables

Example `.env`:

```env
DB_HOST=postgres
DB_PORT=2050
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=techtest_indico

GIN_MODE=release
```

---

## 🚀 Running Locally (Without Docker)

### 1. Setup database

You need PostgreSQL running locally:

```bash
createdb techtest_indico
```

### 2. Run migrations

Using `golang-migrate` or your preferred tool:

```bash
migrate -database "postgres://postgres:postgres@localhost:5432/techtest_indico?sslmode=disable" \
        -path db/migration up
```

### 3. Run application

```bash
go run cmd/api/main.go
```

Server will start at:

```
http://localhost:8080
```

Swagger UI:

```
http://localhost:8080/swagger/index.html
```

---

## 🐳 Running with Docker

### 1. Using docker-compose

Just run:

```bash
docker compose up -d --build
```

Services:

| Service  | Local URL                                | Production URL                                            |
| -------- | ---------------------------------------- | --------------------------------------------------------- |
| Backend  | http://localhost:2051                    | https://techtest-indico-be.alifdwt.com                    |
| Swagger  | http://localhost:2051/swagger/index.html | https://techtest-indico-be.alifdwt.com/swagger/index.html |
| Postgres | localhost:2050                           | -                                                         |

---

## 📜 API Endpoints Summary

| Method | Endpoint             | Description                  |
| ------ | -------------------- | ---------------------------- |
| POST   | /login               | User login                   |
| GET    | /vouchers            | List vouchers                |
| POST   | /vouchers            | Create voucher               |
| GET    | /vouchers/{id}       | Get voucher by ID            |
| PUT    | /vouchers/{id}       | Update voucher               |
| DELETE | /vouchers/{id}       | Delete voucher               |
| POST   | /vouchers/upload-csv | Bulk upload vouchers via CSV |
| GET    | /vouchers/export     | Export vouchers to CSV       |
| GET    | /health              | Health check                 |

---

## 🧪 Testing CSV Upload

Sample CSV files are available in:

```
/filetest
```

Example format:

```csv
discount_percent,expiry_date,voucher_code
10,2025-01-01,SUCCESS011
50,2025-02-15,SUCCESS012
```

Header order doesn't matter.

---

## 🔐 Authentication

- Login returns JWT token.
- Use in request headers:

```
Authorization: Bearer <token>
```

- All protected endpoints require this header.
- Enforced via `auth_middleware.go`.

---

## 📖 API Documentation

Swagger UI:

```
/swagger/index.html
```

Generated from:

```
/docs/swagger.yaml
```

---

## ✅ Final Notes

This backend is:

- Fully containerized
- Production-ready
- Designed with clean separation of layers
- Integrated with SQLC for safe and maintainable DB access

Feel free to explore or test using Postman / Swagger UI.

Thank you for reviewing this technical test 🙏
