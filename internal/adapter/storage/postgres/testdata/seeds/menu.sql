-- Insert Categories
INSERT INTO product_categories (id, name)
VALUES ('11111111-1111-1111-1111-111111111111', 'Appetizers'),
       ('22222222-2222-2222-2222-222222222222', 'Main Dishes'),
       ('33333333-3333-3333-3333-333333333333', 'Desserts'),
       ('44444444-4444-4444-4444-444444444444', 'Drinks'),
       ('55555555-5555-5555-5555-555555555555', 'Sides'),
       ('66666666-6666-6666-6666-666666666666', 'Salads');

-- Insert Products
-- Appetizers
INSERT INTO products (id, name, description, image_path, category, price)
VALUES ('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1', 'Bruschetta', 'Toasted bread topped with fresh tomatoes and basil.',
        'image1', '11111111-1111-1111-1111-111111111111', 7.50),
       ('a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2', 'Garlic Bread', 'Crispy garlic bread served warm with herbs.',
        'image2', '11111111-1111-1111-1111-111111111111', 4.00),
       ('a3a3a3a3-a3a3-a3a3-a3a3-a3a3a3a3a3a3', 'Fried Calamari', 'Lightly breaded calamari served with lemon aioli.',
        'image3', '11111111-1111-1111-1111-111111111111', 9.90),
       ('a4a4a4a4-a4a4-a4a4-a4a4-a4a4a4a4a4a4', 'Stuffed Mushrooms', 'Baked mushrooms stuffed with cheese and herbs.',
        'image4', '11111111-1111-1111-1111-111111111111', 6.80),
       ('a5a5a5a5-a5a5-a5a5-a5a5-a5a5a5a5a5a5', 'Chicken Wings', 'Spicy chicken wings tossed in house hot sauce.',
        'image5', '11111111-1111-1111-1111-111111111111', 8.50);

-- Main Dishes
INSERT INTO products (id, name, description, image_path, category, price)
VALUES ('b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1', 'Grilled Salmon', 'Fresh salmon grilled and served with vegetables.',
        'image6', '22222222-2222-2222-2222-222222222222', 18.90),
       ('b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2', 'Beef Steak', 'Tender seasoned beef steak cooked to preference.',
        'image7', '22222222-2222-2222-2222-222222222222', 22.50),
       ('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b3', 'Chicken Alfredo', 'Creamy Alfredo pasta topped with grilled chicken.',
        'image8', '22222222-2222-2222-2222-222222222222', 14.70),
       ('b4b4b4b4-b4b4-b4b4-b4b4-b4b4b4b4b4b4', 'Margherita Pizza',
        'Classic pizza with tomatoes, basil, and mozzarella.',
        'image9', '22222222-2222-2222-2222-222222222222', 12.00),
       ('b5b5b5b5-b5b5-b5b5-b5b5-b5b5b5b5b5b5', 'BBQ Ribs', 'Slow cooked pork ribs glazed with BBQ sauce.',
        'image10', '22222222-2222-2222-2222-222222222222', 19.50);

-- Desserts
INSERT INTO products (id, name, description, image_path, category, price)
VALUES ('c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 'Cheesecake', 'Creamy cheesecake with a buttery crust.',
        'image11', '33333333-3333-3333-3333-333333333333', 6.50),
       ('c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 'Chocolate Cake', 'Rich layered chocolate cake with ganache.',
        'image12', '33333333-3333-3333-3333-333333333333', 5.90),
       ('c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 'Tiramisu', 'Classic Italian dessert with mascarpone cream.',
        'image13', '33333333-3333-3333-3333-333333333333', 7.20),
       ('c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4', 'Ice Cream Sundae', 'Vanilla ice cream topped with chocolate syrup.',
        'image14', '33333333-3333-3333-3333-333333333333', 4.20),
       ('c5c5c5c5-c5c5-c5c5-c5c5-c5c5c5c5c5c5', 'Apple Pie', 'Warm apple pie with cinnamon and flaky crust.',
        'image15', '33333333-3333-3333-3333-333333333333', 5.40);

-- Drinks
INSERT INTO products (id, name, description, image_path, category, price)
VALUES ('d1d1d1d1-d1d1-d1d1-d1d1-d1d1d1d1d1d1', 'Coca Cola', 'Classic refreshing cola served chilled.',
        'image16',
        '44444444-4444-4444-4444-444444444444', 2.50),
       ('d2d2d2d2-d2d2-d2d2-d2d2-d2d2d2d2d2d2', 'Orange Juice', 'Freshly squeezed orange juice with pulp.',
        'image17',
        '44444444-4444-4444-4444-444444444444', 3.20),
       ('d3d3d3d3-d3d3-d3d3-d3d3-d3d3d3d3d3d3', 'Iced Tea', 'Cold brewed tea sweetened with lemon.',
        'image18',
        '44444444-4444-4444-4444-444444444444', 2.80),
       ('d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4', 'Latte', 'Hot espresso with steamed milk foam.',
        'image19',
        '44444444-4444-4444-4444-444444444444', 4.50),
       ('d5d5d5d5-d5d5-d5d5-d5d5-d5d5d5d5d5d5', 'Mineral Water', 'Clean refreshing mineral water bottle.',
        'image20',
        '44444444-4444-4444-4444-444444444444', 1.50);

-- Sides
INSERT INTO products (id, name, description, image_path, category, price)
VALUES ('e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1', 'French Fries', 'Crispy golden fries seasoned with salt.',
        'image21',
        '55555555-5555-5555-5555-555555555555', 3.50),
       ('e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2', 'Side Salad', 'Fresh greens with tomatoes and vinaigrette.',
        'image22',
        '55555555-5555-5555-5555-555555555555', 3.80),
       ('e3e3e3e3-e3e3-e3e3-e3e3-e3e3e3e3e3e3', 'Rice Bowl', 'Steamed white rice served hot and fluffy.',
        'image23',
        '55555555-5555-5555-5555-555555555555', 2.90),
       ('e4e4e4e4-e4e4-e4e4-e4e4-e4e4e4e4e4e4', 'Mashed Potatoes', 'Creamy mashed potatoes with butter.',
        'image24',
        '55555555-5555-5555-5555-555555555555', 3.40),
       ('e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5', 'Onion Rings', 'Crispy battered onion rings fried to golden brown.',
        'image25', '55555555-5555-5555-5555-555555555555', 3.60);
