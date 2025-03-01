CREATE TABLE IF NOT EXISTS sources (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'available',
    weight FLOAT NOT NULL,
    uuid UUID NOT NULL,
    parent_uuid UUID,

    UNIQUE (uuid)
)