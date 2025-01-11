CREATE TABLE IF NOT EXISTS accounts(
    tx_min INT NOT NULL,
    tx_max INT NOT NULL DEFAULT 0,
    tx_min_committed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_max_committed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_min_rolled_back BOOLEAN NOT NULL DEFAULT FALSE,
    tx_max_rolled_back BOOLEAN NOT NULL DEFAULT FALSE,
    id INT NOT NULL,
    user_id INT NOT NULL,
    balance INT NOT NULL DEFAULT 0,
    PRIMARY KEY (id, tx_min),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS accounts_id_seq;

CREATE INDEX idx_accounts_version ON accounts(id, tx_min, tx_max);
CREATE INDEX idx_accounts_user ON accounts(user_id) WHERE tx_max = 0;