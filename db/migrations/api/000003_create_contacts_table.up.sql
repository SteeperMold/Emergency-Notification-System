CREATE TABLE IF NOT EXISTS contacts
(
    id         SERIAL PRIMARY KEY,
    user_id    INT REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    phone      TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE (user_id, phone)
);
