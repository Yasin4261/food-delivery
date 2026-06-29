-- Trigram indexes so the search ILIKE '%q%' queries can use an index instead of
-- a full scan. pg_trgm ships with the standard PostgreSQL contrib modules.
create extension if not exists pg_trgm;

create index if not exists idx_chefs_business_name_trgm on chefs using gin (business_name gin_trgm_ops);
create index if not exists idx_menu_items_name_trgm on menu_items using gin (name gin_trgm_ops);
create index if not exists idx_users_username_trgm on users using gin (username gin_trgm_ops);
create index if not exists idx_users_email_trgm on users using gin (email gin_trgm_ops);
