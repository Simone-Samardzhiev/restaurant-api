DO
$$
    BEGIN
        RAISE EXCEPTION 'Cannot migrate down automatically: new product status values are already in use';
    END
$$;