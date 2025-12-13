-- Drop indexes
drop index if exists idx_order_items_chef_id;
drop index if exists idx_order_items_menu_item_id;
drop index if exists idx_order_items_order_id;

-- Drop table
drop table if exists order_items;
