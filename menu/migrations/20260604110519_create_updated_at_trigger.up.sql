CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    IF ROW (NEW.*) IS DISTINCT FROM ROW (OLD.*) THEN
        NEW.updated_at = NOW();
    END IF;
    RETURN NEW;
END ;
$$ language plpgsql;

CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE
    ON categories
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE
    ON products
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();