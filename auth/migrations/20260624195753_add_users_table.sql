CREATE TYPE user_roles AS ENUM ('admin', 'cook', 'waitress', 'client');

CREATE TABLE users
(
    id         UUID PRIMARY KEY,
    name       VARCHAR(128) NOT NULL CHECK ( length(name) > 3 ),
    email      VARCHAR(256) NOT NULL UNIQUE CHECK ( length(email) > 8 ),
    password   VARCHAR(256) NOT NULL,
    role       user_roles   NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL CHECK ( updated_at > created_at )
)