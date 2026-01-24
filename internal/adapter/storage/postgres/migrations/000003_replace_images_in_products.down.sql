ALTER TABLE products
    ADD COLUMN image_url        VARCHAR(200),
    ADD COLUMN delete_image_url VARCHAR(200),
    DROP COLUMN image_path;
