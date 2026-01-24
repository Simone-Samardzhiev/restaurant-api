ALTER TABLE products
    DROP COLUMN image_url,
    DROP COLUMN delete_image_url,
    ADD COLUMN image_path VARCHAR(255) NOT NULL UNIQUE;

