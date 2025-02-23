CREATE TABLE IF NOT EXISTS sources (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'available',
    weight FLOAT NOT NULL
    uuid uuid NOT NULL
    parent_uuid uuid

    UNIQUE (uuid)
)