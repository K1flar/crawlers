CREATE TABLE IF NOT EXISTS tasks_x_sources (
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    launch_id BIGINT NOT NULL REFERENCES launches(id) ON DELETE CASCADE,
    source_id BIGINT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    parent_source_id BIGINT REFERENCES sources(id) ON DELETE CASCADE, 
    weight FLOAT NOT NULL,

    PRIMARY KEY (task_id, launch_id, source_id)
)