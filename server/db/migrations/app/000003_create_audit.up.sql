CREATE TABLE IF NOT EXISTS audit(
    tx_min INT NOT NULL,
    tx_max INT NOT NULL DEFAULT 0,
    tx_min_committed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_max_committed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_min_rolled_back BOOLEAN NOT NULL DEFAULT FALSE,
    tx_max_rolled_back BOOLEAN NOT NULL DEFAULT FALSE,
    id INT NOT NULL,  -- Remove SERIAL, we'll manage IDs
    timestamp TIMESTAMP NOT NULL,
    operation TEXT NOT NULL,
    user_id INT NOT NULL,
    PRIMARY KEY (id, tx_min),  -- Composite key for versioning
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE SEQUENCE IF NOT EXISTS audit_id_seq;

CREATE INDEX idx_audit_version ON audit(id, tx_min, tx_max);