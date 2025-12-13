create table if not exists chefs (
    id serial primary key,
    user_id integer unique not null references users(id) on delete cascade,
    
    -- Chef profile information
    business_name varchar(200) not null,
    bio text,
    specialty varchar(100),
    experience_years integer,
    
    -- Location
    kitchen_address text not null,
    kitchen_city varchar(100),
    kitchen_latitude decimal(10, 8),
    kitchen_longitude decimal(11, 8),
    delivery_radius integer default 5, -- in kilometers
    
    -- Certificates and verification
    food_license_number varchar(100),
    health_certificate_url varchar(500),
    is_verified boolean default false,
    verified_at timestamp,
    
    -- Statistics
    rating decimal(3, 2) default 0.00, -- 0.00 - 5.00
    total_reviews integer default 0,
    total_orders integer default 0,
    
    -- Status
    is_active boolean default true,
    is_accepting_orders boolean default true,
    
    -- Timestamps
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

-- Indexes
create index if not exists idx_chefs_user_id on chefs(user_id);
create index if not exists idx_chefs_is_verified on chefs(is_verified);
create index if not exists idx_chefs_is_active on chefs(is_active);
create index if not exists idx_chefs_rating on chefs(rating);
create index if not exists idx_chefs_location on chefs(kitchen_latitude, kitchen_longitude);
create index if not exists idx_chefs_city on chefs(kitchen_city);

-- Comments
comment on table chefs is 'Chef/home cook profiles';
comment on column chefs.is_verified is 'Admin verification status';
comment on column chefs.delivery_radius is 'Maximum delivery distance in kilometers';
comment on column chefs.is_accepting_orders is 'Whether chef is currently accepting new orders';
