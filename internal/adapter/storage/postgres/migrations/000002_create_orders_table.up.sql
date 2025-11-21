CREATE TYPE order_status AS ENUM ('closed', 'open', 'paid');

CREATE TABLE orders
(
    id           UUID PRIMARY KEY,
    table_number INT          NOT NULL CHECK ( table_number > 0 ),
    status       order_status NOT NULL
);


CREATE TYPE ordered_product_status AS ENUM ('pending', 'preparing', 'done');
CREATE TABLE ordered_products
(
    id         UUID PRIMARY KEY,
    product_id UUID                   NOT NULL REFERENCES products (id),
    status     ordered_product_status NOT NULL,
    order_id   UUID                   NOT NULL REFERENCES orders (id)
)