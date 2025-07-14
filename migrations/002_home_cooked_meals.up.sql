-- +goose Up
-- Update users table for home-cooked meal platform
ALTER TABLE users 
ADD COLUMN phone VARCHAR(20),
ADD COLUMN is_active BOOLEAN DEFAULT true;

-- Update role column to include chef
ALTER TABLE users 
ALTER COLUMN role SET DEFAULT 'customer';

-- Create chefs table
CREATE TABLE chefs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    kitchen_name VARCHAR(100) NOT NULL,
    description TEXT,
    speciality VARCHAR(100),
    experience INTEGER DEFAULT 0,
    address TEXT NOT NULL,
    district VARCHAR(50),
    city VARCHAR(50),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    rating DECIMAL(3,2) DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create meals table (replacing products for home-cooked meals)
CREATE TABLE meals (
    id SERIAL PRIMARY KEY,
    chef_id INTEGER REFERENCES chefs(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    cuisine VARCHAR(50),
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'TRY',
    portion VARCHAR(50),
    serving_size INTEGER DEFAULT 1,
    available_quantity INTEGER DEFAULT 0,
    preparation_time INTEGER,
    cooking_time INTEGER,
    calories INTEGER,
    ingredients TEXT,
    allergens TEXT,
    is_vegetarian BOOLEAN DEFAULT false,
    is_vegan BOOLEAN DEFAULT false,
    is_gluten_free BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    is_available BOOLEAN DEFAULT true,
    rating DECIMAL(3,2) DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    images TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Update cart_items table for meals
ALTER TABLE cart_items 
ADD COLUMN meal_id INTEGER REFERENCES meals(id),
ADD COLUMN chef_id INTEGER REFERENCES chefs(id),
ADD COLUMN special_instructions TEXT;

-- Update orders table for home-cooked meal delivery
ALTER TABLE orders 
ADD COLUMN order_number VARCHAR(50) UNIQUE,
ADD COLUMN currency VARCHAR(3) DEFAULT 'TRY',
ADD COLUMN delivery_type VARCHAR(20),
ADD COLUMN delivery_address TEXT,
ADD COLUMN delivery_date TIMESTAMP,
ADD COLUMN delivery_time VARCHAR(10),
ADD COLUMN payment_method VARCHAR(20),
ADD COLUMN payment_status VARCHAR(20) DEFAULT 'pending',
ADD COLUMN customer_note TEXT,
ADD COLUMN chef_note TEXT;

-- Update order_items table for meals
ALTER TABLE order_items 
ADD COLUMN meal_id INTEGER REFERENCES meals(id),
ADD COLUMN chef_id INTEGER REFERENCES chefs(id),
ADD COLUMN subtotal DECIMAL(10,2),
ADD COLUMN special_instructions TEXT;

-- Create reviews table
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    chef_id INTEGER REFERENCES chefs(id),
    meal_id INTEGER REFERENCES meals(id),
    order_id INTEGER REFERENCES orders(id) NOT NULL,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    title VARCHAR(100),
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for new tables
CREATE INDEX idx_chefs_user ON chefs(user_id);
CREATE INDEX idx_chefs_location ON chefs(city, district);
CREATE INDEX idx_chefs_active ON chefs(is_active);
CREATE INDEX idx_meals_chef ON meals(chef_id);
CREATE INDEX idx_meals_category ON meals(category);
CREATE INDEX idx_meals_available ON meals(is_active, is_available);
CREATE INDEX idx_cart_items_meal ON cart_items(meal_id);
CREATE INDEX idx_order_items_meal ON order_items(meal_id);
CREATE INDEX idx_reviews_chef ON reviews(chef_id);
CREATE INDEX idx_reviews_meal ON reviews(meal_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);

-- +goose Down
-- Remove new columns and tables
DROP INDEX IF EXISTS idx_reviews_user;
DROP INDEX IF EXISTS idx_reviews_meal;
DROP INDEX IF EXISTS idx_reviews_chef;
DROP INDEX IF EXISTS idx_order_items_meal;
DROP INDEX IF EXISTS idx_cart_items_meal;
DROP INDEX IF EXISTS idx_meals_available;
DROP INDEX IF EXISTS idx_meals_category;
DROP INDEX IF EXISTS idx_meals_chef;
DROP INDEX IF EXISTS idx_chefs_active;
DROP INDEX IF EXISTS idx_chefs_location;
DROP INDEX IF EXISTS idx_chefs_user;

DROP TABLE IF EXISTS reviews;

ALTER TABLE order_items 
DROP COLUMN IF EXISTS special_instructions,
DROP COLUMN IF EXISTS subtotal,
DROP COLUMN IF EXISTS chef_id,
DROP COLUMN IF EXISTS meal_id;

ALTER TABLE orders 
DROP COLUMN IF EXISTS chef_note,
DROP COLUMN IF EXISTS customer_note,
DROP COLUMN IF EXISTS payment_status,
DROP COLUMN IF EXISTS payment_method,
DROP COLUMN IF EXISTS delivery_time,
DROP COLUMN IF EXISTS delivery_date,
DROP COLUMN IF EXISTS delivery_address,
DROP COLUMN IF EXISTS delivery_type,
DROP COLUMN IF EXISTS currency,
DROP COLUMN IF EXISTS order_number;

ALTER TABLE cart_items 
DROP COLUMN IF EXISTS special_instructions,
DROP COLUMN IF EXISTS chef_id,
DROP COLUMN IF EXISTS meal_id;

DROP TABLE IF EXISTS meals;
DROP TABLE IF EXISTS chefs;

ALTER TABLE users 
DROP COLUMN IF EXISTS is_active,
DROP COLUMN IF EXISTS phone;
