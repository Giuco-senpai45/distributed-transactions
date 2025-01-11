CREATE TABLE IF NOT EXISTS locks(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    shared BOOLEAN NOT NULL DEFAULT FALSE,
    record_table TEXT NOT NULL,
    record_id INT NOT NULL,
    txid INT NOT NULL REFERENCES transactions(id)
);