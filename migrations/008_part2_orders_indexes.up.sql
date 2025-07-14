-- Part 2: Create indexes for orders table
-- This part creates indexes for the orders table

-- Create indexes for location-based queries
CREATE INDEX IF NOT EXISTS idx_orders_delivery_location ON orders(delivery_latitude, delivery_longitude);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
