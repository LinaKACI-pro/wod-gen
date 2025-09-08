CREATE TABLE wods (
    id UUID PRIMARY KEY,
    seed TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    level TEXT NOT NULL,
    duration_min INT NOT NULL,
    equipment TEXT[] NOT NULL,
    blocks JSONB NOT NULL
);

CREATE INDEX idx_wods_seed_level_duration
    ON wods(seed, level, duration_min);