-- +goose Up
CREATE TYPE gender_enum AS ENUM ('M', 'F');

ALTER TABLE users
    ADD COLUMN gender gender_enum NOT NULL,
    ADD COLUMN birth_date DATE NOT NULL;

INSERT INTO users (name, email, gender, birth_date) VALUES
('Gleb', 'bolshoyhuy@gmail.com', 'M', '2006-06-16'),
('Sanya', 'tostiykitaec@gmail.com', 'M', '2005-07-30'),
('Islam', 'uzkoglaziy@gmail.com', 'M', '2006-07-04');


-- +goose Down
ALTER TABLE users
DROP COLUMN birth_date,
DROP COLUMN gender;

DROP TYPE gender_enum;