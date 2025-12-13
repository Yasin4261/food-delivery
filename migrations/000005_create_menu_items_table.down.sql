-- Drop indexes
drop index if exists idx_menu_items_is_available;
drop index if exists idx_menu_items_price;
drop index if exists idx_menu_items_rating;
drop index if exists idx_menu_items_is_featured;
drop index if exists idx_menu_items_is_active;
drop index if exists idx_menu_items_cuisine;
drop index if exists idx_menu_items_category;
drop index if exists idx_menu_items_chef_id;
drop index if exists idx_menu_items_menu_id;

-- Drop table
drop table if exists menu_items;
