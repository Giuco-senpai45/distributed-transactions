CREATE TABLE IF NOT EXISTS users(
    tx_min INT NOT NULL,
    tx_max INT NOT NULL DEFAULT 0,
    tx_min_committed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_max_committed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_min_rolled_back BOOLEAN NOT NULL DEFAULT FALSE,
    tx_max_rolled_back BOOLEAN NOT NULL DEFAULT FALSE,
    id INT NOT NULL, 
    username VARCHAR (100) NOT NULL,
    PRIMARY KEY (id, tx_min, username),
    UNIQUE (id) 
);

CREATE SEQUENCE IF NOT EXISTS users_id_seq;

CREATE INDEX idx_users_version ON users(id, tx_min, tx_max, username);