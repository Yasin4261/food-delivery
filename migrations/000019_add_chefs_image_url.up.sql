-- Kitchen photo for the chef storefront (#63). Dishes already carry
-- menu_items.image_url since 000004.
alter table chefs add column if not exists image_url text;
