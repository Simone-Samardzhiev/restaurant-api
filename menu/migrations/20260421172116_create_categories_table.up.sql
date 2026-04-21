CREATE TABLE categories
(
    id         UUID PRIMARY KEY,
    name       VARCHAR(64) NOT NULL UNIQUE CHECK ( length(name) > 2 ),
    created_at TIMESTAMPTZ NULL NULL,
    updated_at TIMESTAMPTZ NULL NULL CHECK ( updated_at >= created_at )
);
