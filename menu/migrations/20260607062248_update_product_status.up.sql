CREATE TYPE product_status_new AS ENUM ('ready', 'missing_image', 'awaiting_image_update');

BEGIN;
ALTER TABLE products
    ALTER COLUMN status TYPE product_status_new USING (
        CASE status::text
            WHEN 'awaiting_image' THEN 'missing_image'::product_status_new
            ELSE status::text::product_status_new END
        );

ALTER TYPE product_status RENAME TO product_status_old;

DROP TYPE product_status_old;

ALTER TYPE product_status_new RENAME TO product_status;
COMMIT;