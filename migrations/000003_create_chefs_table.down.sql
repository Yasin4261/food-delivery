-- Drop indexes
drop index if exists idx_chefs_city;
drop index if exists idx_chefs_location;
drop index if exists idx_chefs_rating;
drop index if exists idx_chefs_is_active;
drop index if exists idx_chefs_is_verified;
drop index if exists idx_chefs_user_id;

-- Drop table
drop table if exists chefs;
