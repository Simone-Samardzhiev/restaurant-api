ALTER TABLE products
    DROP COLUMN IF EXISTS pending_image_key,
    DROP COLUMN IF EXISTS pending_image_type;