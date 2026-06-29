drop index if exists idx_chefs_is_online;
alter table chefs drop column if exists is_online;
