CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,
    depth_level INT NOT NULL,
    min_weight FLOAT NOT NULL,
    max_sources INT NOT NULL,
    max_neighbours_for_source INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_query ON tasks (query text_pattern_ops);