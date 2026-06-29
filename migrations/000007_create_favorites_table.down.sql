-- Drop indexes
drop index if exists idx_favorites_chef_id;
drop index if exists idx_favorites_user_id;

-- Drop table
drop table if exists favorites;
