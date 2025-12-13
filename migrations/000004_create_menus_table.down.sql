-- Drop indexes
drop index if exists idx_menus_menu_type;
drop index if exists idx_menus_is_featured;
drop index if exists idx_menus_is_active;
drop index if exists idx_menus_chef_id;

-- Drop table
drop table if exists menus;
