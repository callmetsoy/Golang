
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS users;


CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(100) NOT NULL,
                       email VARCHAR(100) UNIQUE NOT NULL,
                       password TEXT NOT NULL,
                       age INT,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       deleted_at TIMESTAMP
);


CREATE TABLE audit_logs (
                            id SERIAL PRIMARY KEY,
                            user_id INT,
                            action VARCHAR(50),
                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);