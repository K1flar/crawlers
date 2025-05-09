CREATE TABLE IF NOT EXISTS launches (
    id BIGSERIAL PRIMARY KEY,
    number BIGINT NOT NULL,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    sources_viewed INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    error VARCHAR(30),

    UNIQUE (number, task_id)
);