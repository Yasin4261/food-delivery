-- Drop indexes
drop index if exists idx_orders_created_at;
drop index if exists idx_orders_payment_status;
drop index if exists idx_orders_status;
drop index if exists idx_orders_user_id;
drop index if exists idx_orders_order_code;

-- Drop table
drop table if exists orders;