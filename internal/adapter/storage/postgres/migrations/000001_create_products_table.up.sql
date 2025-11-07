CREATE TABLE products_categories
(
    id   UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE CHECK ( length(name) >= 4 )
);

CREATE TABLE products
(
    id          UUID PRIMARY KEY,
    name        VARCHAR(100) UNIQUE                      NOT NULL CHECK ( length(name) >= 3 ),
    description TEXT                                     NOT NULL CHECK ( length(description) >= 15 ),
    image_path  VARCHAR(200)                             NULL CHECK ( length(description) >= 10 ),
    category    UUID REFERENCES products_categories (id) NOT NULL,
    price       DECIMAL(8, 2)                            NOT NULL CHECK ( price > 0 )
);

