-- +goose Up
CREATE TABLE user_friends (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  friend_id UUID REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, friend_id)
);

INSERT INTO users (name, email, gender, birth_date) VALUES
('User1','u1@mail.com','M','1995-01-01'),
('User2','u2@mail.com','F','1996-02-01'),
('User3','u3@mail.com','M','1997-03-01'),
('User4','u4@mail.com','F','1998-04-01'),
('User5','u5@mail.com','M','1994-05-01'),
('User6','u6@mail.com','F','1993-06-01'),
('User7','u7@mail.com','M','1992-07-01'),
('User8','u8@mail.com','F','1991-08-01'),
('User9','u9@mail.com','M','1990-09-01'),
('User10','u10@mail.com','F','1995-10-01'),
('User11','u11@mail.com','M','1996-11-01'),
('User12','u12@mail.com','F','1997-12-01'),
('User13','u13@mail.com','M','1998-01-10'),
('User14','u14@mail.com','F','1999-02-10'),
('User15','u15@mail.com','M','1993-03-10'),
('User16','u16@mail.com','F','1994-04-10'),
('User17','u17@mail.com','M','1992-05-10'),
('User18','u18@mail.com','F','1991-06-10'),
('User19','u19@mail.com','M','1990-07-10'),
('User20','u20@mail.com','F','1995-08-10');

INSERT INTO user_friends (user_id, friend_id)
SELECT u1.id, u2.id
FROM users u1, users u2
WHERE u1.name = 'User1' AND u2.name = 'User3';

INSERT INTO user_friends (user_id, friend_id)
SELECT u1.id, u2.id
FROM users u1, users u2
WHERE u1.name = 'User1' AND u2.name = 'User4';

INSERT INTO user_friends (user_id, friend_id)
SELECT u1.id, u2.id
FROM users u1, users u2
WHERE u1.name = 'User1' AND u2.name = 'User5';

INSERT INTO user_friends (user_id, friend_id)
SELECT u1.id, u2.id
FROM users u1, users u2
WHERE u1.name = 'User2' AND u2.name = 'User3';

INSERT INTO user_friends (user_id, friend_id)
SELECT u1.id, u2.id
FROM users u1, users u2
WHERE u1.name = 'User2' AND u2.name = 'User4';

INSERT INTO user_friends (user_id, friend_id)
SELECT u1.id, u2.id
FROM users u1, users u2
WHERE u1.name = 'User2' AND u2.name = 'User5';

-- +goose Down
DROP TABLE user_friends;
