-- Part 3: Create sub_orders table
-- This part creates the new sub_orders table

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
