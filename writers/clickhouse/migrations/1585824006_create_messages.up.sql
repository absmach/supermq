CREATE TABLE
IF NOT EXISTS messages
(
    id            UUID,
    channel       UUID,
    subtopic      String,
    publisher     UUID,
    protocol      String,
    name          String,
    unit          String,
    value         Nullable(Float64),
    string_value  Nullable(String),
    bool_value    Nullable(UInt8),
    data_value    Nullable(String),
    sum           Nullable(Float64),
    time          DateTime,
    update_time   Float64
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(time)
ORDER BY (channel, time)
