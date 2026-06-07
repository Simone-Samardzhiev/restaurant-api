ALTER TABLE products
    ADD COLUMN pending_image_key          TEXT UNIQUE,
    ADD COLUMN pending_image_content_type product_image_content_type;
