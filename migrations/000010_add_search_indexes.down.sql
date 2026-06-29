drop index if exists idx_users_email_trgm;
drop index if exists idx_users_username_trgm;
drop index if exists idx_menu_items_name_trgm;
drop index if exists idx_chefs_business_name_trgm;
drop extension if exists pg_trgm;
