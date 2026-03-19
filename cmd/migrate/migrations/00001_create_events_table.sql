-- +goose Up
CREATE TABLE
    IF NOT EXISTS events (
        event_name String,
        channel String,
        campaign_id String,
        user_id String,
        timestamp Int64,
        event_time DateTime MATERIALIZED toDateTime (timestamp),
        tags Array (String),
        metadata String
    ) ENGINE = MergeTree
PARTITION BY
    toYYYYMM (event_time)
ORDER BY
    (event_name, event_time, channel, user_id);

-- +goose Down
DROP TABLE IF EXISTS events;