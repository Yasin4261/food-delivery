-- Migration for multi-vendor order support
-- This migration converts the single-chef order system to multi-vendor support

-- First, let's update the orders table structure
ALTER TABLE orders DROP COLUMN IF EXISTS chef_id;

-- Add new columns for multi-vendor support
ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_number VARCHAR(50) UNIQUE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'TRY';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_type VARCHAR(20) DEFAULT 'delivery';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_address TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_latitude DECIMAL(10,8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_longitude DECIMAL(11,8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_radius DECIMAL(5,2) DEFAULT 10;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_note TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS chef_count INTEGER DEFAULT 0;

-- Create indexes for location-based queries
CREATE INDEX IF NOT EXISTS idx_orders_delivery_location ON orders(delivery_latitude, delivery_longitude);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- Create sub_orders table for multi-vendor support
CREATE TABLE IF NOT EXISTS sub_orders (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    chef_id INTEGER NOT NULL REFERENCES chefs(id) ON DELETE CASCADE,
    chef_name VARCHAR(255) NOT NULL,
    chef_business_name VARCHAR(255),
    subtotal DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    chef_status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected, preparing, ready, delivered
    preparation_time INTEGER, -- minutes
    estimated_ready_time TIMESTAMP,
    chef_note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sub_orders
CREATE INDEX IF NOT EXISTS idx_sub_orders_order_id ON sub_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_sub_orders_chef_id ON sub_orders(chef_id);
CREATE INDEX IF NOT EXISTS idx_sub_orders_status ON sub_orders(status);
CREATE INDEX IF NOT EXISTS idx_sub_orders_chef_status ON sub_orders(chef_status);

-- Update order_items table to remove chef_id since it's now handled by sub_orders
ALTER TABLE order_items DROP COLUMN IF EXISTS chef_id;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS sub_order_id INTEGER REFERENCES sub_orders(id) ON DELETE CASCADE;

-- Create index for order_items sub_order_id
CREATE INDEX IF NOT EXISTS idx_order_items_sub_order_id ON order_items(sub_order_id);

-- Update reviews table to support both meal and chef reviews
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE;
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS sub_order_id INTEGER REFERENCES sub_orders(id) ON DELETE CASCADE;
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS review_type VARCHAR(20) DEFAULT 'meal'; -- meal, chef, overall

-- Generate order numbers for existing orders (if any)
UPDATE orders SET order_number = 'ORD-' || to_char(created_at, 'YYYYMMDD') || '-' || LPAD(id::text, 3, '0') 
WHERE order_number IS NULL;

-- Add constraint to ensure order_number is not null for new records
ALTER TABLE orders ALTER COLUMN order_number SET NOT NULL;

-- Add check constraints
ALTER TABLE orders ADD CONSTRAINT chk_orders_status 
    CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'delivered', 'cancelled'));

ALTER TABLE orders ADD CONSTRAINT chk_orders_payment_status 
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded'));

ALTER TABLE orders ADD CONSTRAINT chk_orders_delivery_type 
    CHECK (delivery_type IN ('pickup', 'delivery'));

ALTER TABLE sub_orders ADD CONSTRAINT chk_sub_orders_status 
    CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'delivered', 'cancelled'));

ALTER TABLE sub_orders ADD CONSTRAINT chk_sub_orders_chef_status 
    CHECK (chef_status IN ('pending', 'accepted', 'rejected', 'preparing', 'ready', 'delivered'));

ALTER TABLE reviews ADD CONSTRAINT chk_reviews_review_type 
    CHECK (review_type IN ('meal', 'chef', 'overall'));
