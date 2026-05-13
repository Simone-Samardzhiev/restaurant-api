CREATE TYPE product_status AS ENUM ('ready','awaiting_image');

CREATE TABLE products
(
    id          UUID PRIMARY KEY,
    name        VARCHAR(128)   NOT NULL CHECK ( length(name) > 3),
    description TEXT           NOT NULL CHECK ( length(description) > 16 ),
    price       DECIMAL(6, 2)  NOT NULL CHECK ( price > 0 ),
    category_id UUID           NOT NULL REFERENCES categories (id),

    image_key   TEXT           NOT NULL UNIQUE,
    status      product_status NOT NULL,

    created_at  TIMESTAMPTZ    NOT NULL,
    updated_at  TIMESTAMPTZ    NOT NULL CHECK ( updated_at >= created_at )
)