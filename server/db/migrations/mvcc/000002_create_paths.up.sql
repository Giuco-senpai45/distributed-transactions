CREATE TABLE IF NOT EXISTS paths(
    id SERIAL PRIMARY KEY,
    path ltree NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    dependency_type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX path_idx ON paths USING GIST (path);