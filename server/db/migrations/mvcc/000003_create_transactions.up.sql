CREATE TABLE IF NOT EXISTS transactions(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    status TEXT,
    timestamp BIGINT NOT NULL
);