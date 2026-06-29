-- Drop indexes
drop index if exists idx_reviews_order_id;
drop index if exists idx_reviews_user_id;
drop index if exists idx_reviews_menu_item_id;
drop index if exists idx_reviews_chef_id;

-- Drop table
drop table if exists reviews;
