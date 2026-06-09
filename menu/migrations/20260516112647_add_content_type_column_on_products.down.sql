ALTER TABLE products
    DROP COLUMN IF EXISTS image_content_type;

DROP DOMAIN IF EXISTS image_content_type;