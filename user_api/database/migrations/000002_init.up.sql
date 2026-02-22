CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    age INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
    );

INSERT INTO users (name, email, age, password)
VALUES ('John Doe', 'john@example.com', 30, 'temporary_password');