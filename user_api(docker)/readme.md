# 🚀 User Management API (Golang & Docker)

Production-ready RESTful API built with **Go** and **PostgreSQL**, now fully containerized with **Docker** and **Multi-stage builds**.

---

## 🏗 Architecture & Design

The system follows a strict layered approach to separate concerns:

* **Delivery (Handlers)**: Managed via `net/http`, handles routing and JSON parsing.
* **App (Usecase)**: Core business logic and password hashing.
* **Repository (Postgres)**: Data access layer using `sqlx`. Implements **ACID Transactions** for atomic user updates and audit logging.

---

## 🐳 Docker Features (Assignment #4)

* **Multi-stage Build**: Optimized Docker image (~28MB) using `golang:alpine` for building and `alpine` for production.
* **Automated Init**: Database schema is automatically created on first start via `init.sql`.
* **Healthchecks**: The API service waits for PostgreSQL to be "Healthy" before starting.
* **Data Persistence**: Uses **Named Volumes** to ensure data survives container restarts.
* **Environment Isolation**: Fully configured via `docker-compose.yml`.

---

## 🛠 Tech Stack

* **Language:** Go 1.21+
* **Database:** PostgreSQL 15 (Alpine)
* **Containerization:** Docker & Docker Compose
* **Libraries:** `sqlx`, `bcrypt`, `pq`, `godotenv`.

---

## 📥 Quick Start (Docker)

The fastest way to run the project is using Docker Compose. No manual DB setup required.

```bash
# 1. Clone and enter the directory
git clone https://github.com/callmetsoy/Golang.git
cd user_api

# 2. Start the entire stack
docker compose up --build

```

*The API will be available at `http://localhost:8080`.*

---

## 🔌 API Endpoints

| Method | Endpoint | Description | Auth Required |
| --- | --- | --- | --- |
| `GET` | `/health` | Service health status | No |
| `GET` | `/users` | Fetch all active users (Soft Delete aware) | **Yes** |
| `POST` | `/users` | Create user + Audit Log | **Yes** |
| `DELETE` | `/users/{id}` | Soft delete user + Audit Log | **Yes** |

**Note:** All protected routes require the `X-API-KEY` header.

---

## 📜 Audit & Persistence

Every mutation (Create, Update, Delete) triggers a record in the `audit_logs` table. To verify data persistence after a restart:

```bash
docker compose down
docker compose up
# Your data is still there!

```
