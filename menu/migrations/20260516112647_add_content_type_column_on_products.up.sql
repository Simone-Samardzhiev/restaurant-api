CREATE DOMAIN image_content_type AS TEXT
    CONSTRAINT supported CHECK (
        VALUE IN ('image/jpeg', 'image/png', 'image/webp')
        );


BEGIN;

ALTER TABLE products
    ADD column image_content_type image_content_type;

UPDATE products
SET image_content_type =
        CASE
            WHEN image_key LIKE '%.jpeg' THEN 'image/jpeg'
            WHEN image_key LIKE '%.png' THEN 'image/png'
            WHEN image_key LIKE '%.webp' THEN 'image/webp'
            END;

ALTER TABLE products
    ALTER COLUMN image_content_type SET NOT NULL;

COMMIT;
