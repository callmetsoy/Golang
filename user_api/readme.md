
# 🚀 User Management API (Golang)

Production-ready RESTful API built with **Go** and **PostgreSQL**, implementing **Clean Architecture** for maximum maintainability and scalability.

---

## 🏗 Architecture & Design
The system follows a strict layered approach to separate concerns:

* **Delivery (Handlers)**: Managed via `net/http`, handles routing, JSON parsing, and response status codes.
* **App (Usecase)**: The core business logic layer. Orchestrates data flow and handles security tasks like password hashing.
* **Repository (Postgres)**: Data access layer using `sqlx`. Implements **ACID Transactions** to ensure data integrity between user updates and audit logs.



---

## 🚀 Key Features
* **🛡 Secure Auth**: Custom middleware for API Key authentication via the `X-API-KEY` header.
* **🔑 Password Security**: Industry-standard encryption using `bcrypt`.
* **📜 Audit Trail**: Every mutation (Create, Update, Delete) automatically triggers a record in the `audit_logs` table.
* **🗑 Soft Delete**: Implements safe deletion logic using `deleted_at` timestamps to preserve data history.
* **🛑 Graceful Shutdown**: Listens for `SIGINT`/`SIGTERM` to safely close DB connections and stop the server without data loss.

---

## 🛠 Tech Stack
* **Language:** Go 1.21+
* **Database:** PostgreSQL
* **Libraries:**
    * `sqlx` (Advanced SQL features)
    * `godotenv` (Environment configuration)
    * `bcrypt` (Cryptography)
    * `pq` (Postgres driver)

---

## 📥 Installation & Setup

### 1. Database Schema
Run the following SQL to prepare your database:
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    password TEXT,
    age INT,
    deleted_at TIMESTAMP
);

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

```

### 2. Configuration

Create a `.env` file in the root directory:

```env
DB_HOST=1.1.1.1
DB_PORT=5432
DB_USER=gopher
DB_PASSWORD=
DB_NAME=mydb
DB_SSLMODE=disable
API_KEY=my-secret-key

```

### 3. Execution

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/api/main.go

```

---

## 🔌 API Endpoints

| Method | Endpoint | Description | Auth Required |
| --- | --- | --- | --- |
| `GET` | `/health` | Service health status | No |
| `GET` | `/users` | Fetch all active users | **Yes** |
| `GET` | `/users/{id}` | Fetch user by ID | **Yes** |
| `POST` | `/users` | Create user (hashes password) | **Yes** |
| `PUT` | `/users` | Update user details | **Yes** |
| `DELETE` | `/users/{id}` | Soft delete user | **Yes** |

### Example Request (Delete)

```bash
curl -X DELETE http://localhost:8080/users/1 \
     -H "X-API-KEY: my-secret-key"
