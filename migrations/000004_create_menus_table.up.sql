create table if not exists menus (
    id serial primary key,
    chef_id integer not null references chefs(id) on delete cascade,
    
    -- Menu information
    name varchar(200) not null,
    description text,
    menu_type varchar(50) default 'regular', -- regular, daily_special, seasonal, weekend
    
    -- Scheduling
    available_days varchar(50), -- JSON or CSV: "monday,tuesday,friday"
    available_from time, -- 09:00
    available_until time, -- 22:00
    
    -- Status
    is_active boolean default true,
    is_featured boolean default false,
    
    -- Timestamps
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

-- Indexes
create index if not exists idx_menus_chef_id on menus(chef_id);
create index if not exists idx_menus_is_active on menus(is_active);
create index if not exists idx_menus_is_featured on menus(is_featured);
create index if not exists idx_menus_menu_type on menus(menu_type);

-- Comments
comment on table menus is 'Chef menus and meal collections';
comment on column menus.menu_type is 'Types: regular, daily_special, seasonal, weekend';
comment on column menus.available_days is 'Days when menu is available (comma-separated)';
