-- Migration rollback for multi-vendor order support

-- Drop constraint checks
ALTER TABLE reviews DROP CONSTRAINT IF EXISTS chk_reviews_review_type;
ALTER TABLE sub_orders DROP CONSTRAINT IF EXISTS chk_sub_orders_chef_status;
ALTER TABLE sub_orders DROP CONSTRAINT IF EXISTS chk_sub_orders_status;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_delivery_type;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_payment_status;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_status;

-- Remove NOT NULL constraint from order_number
ALTER TABLE orders ALTER COLUMN order_number DROP NOT NULL;

-- Remove new columns from reviews table
ALTER TABLE reviews DROP COLUMN IF EXISTS review_type;
ALTER TABLE reviews DROP COLUMN IF EXISTS sub_order_id;
ALTER TABLE reviews DROP COLUMN IF EXISTS order_id;

-- Drop indexes for order_items
DROP INDEX IF EXISTS idx_order_items_sub_order_id;

-- Remove sub_order_id from order_items and add back chef_id
ALTER TABLE order_items DROP COLUMN IF EXISTS sub_order_id;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS chef_id INTEGER REFERENCES chefs(id) ON DELETE CASCADE;

-- Drop sub_orders table and its indexes
DROP INDEX IF EXISTS idx_sub_orders_chef_status;
DROP INDEX IF EXISTS idx_sub_orders_status;
DROP INDEX IF EXISTS idx_sub_orders_chef_id;
DROP INDEX IF EXISTS idx_sub_orders_order_id;
DROP TABLE IF EXISTS sub_orders;

-- Drop indexes for orders
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_delivery_location;

-- Remove new columns from orders table
ALTER TABLE orders DROP COLUMN IF EXISTS chef_count;
ALTER TABLE orders DROP COLUMN IF EXISTS customer_note;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_status;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_method;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_radius;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_longitude;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_latitude;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_address;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_type;
ALTER TABLE orders DROP COLUMN IF EXISTS currency;
ALTER TABLE orders DROP COLUMN IF EXISTS order_number;

-- Add back chef_id to orders table
ALTER TABLE orders ADD COLUMN IF NOT EXISTS chef_id INTEGER REFERENCES chefs(id) ON DELETE CASCADE;
