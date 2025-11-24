CREATE TYPE order_session_status AS ENUM ('closed', 'open', 'paid');

CREATE TABLE order_sessions
(
    id           UUID PRIMARY KEY,
    table_number INT                  NOT NULL CHECK ( table_number > 0 ),
    status       order_session_status NOT NULL
);


CREATE TYPE ordered_product_status AS ENUM ('pending', 'preparing', 'done');
CREATE TABLE ordered_products
(
    id         UUID PRIMARY KEY,
    product_id UUID                   NOT NULL REFERENCES products (id),
    status     ordered_product_status NOT NULL,
    session_id UUID                   NOT NULL REFERENCES order_sessions (id)
)