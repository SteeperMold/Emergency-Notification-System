CREATE TABLE IF NOT EXISTS contacts
(
    id         SERIAL PRIMARY KEY,
    user_id    INT REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT NOT NULL UNIQUE,
    phone      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
