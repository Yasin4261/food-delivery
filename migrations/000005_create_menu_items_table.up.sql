create table if not exists menu_items (
    id serial primary key,
    menu_id integer not null references menus(id) on delete cascade,
    chef_id integer not null references chefs(id) on delete cascade,
    
    -- Meal information
    name varchar(200) not null,
    description text,
    category varchar(50), -- appetizer, main_course, dessert, beverage, soup
    cuisine varchar(50), -- turkish, italian, chinese, mexican, etc.
    
    -- Pricing
    price decimal(10, 2) not null,
    original_price decimal(10, 2), -- Original price if discounted
    
    -- Portion and preparation
    portion_size varchar(50), -- 1 person, 2 people, family size
    preparation_time integer, -- in minutes
    serving_size integer default 1,
    
    -- Stock
    available_quantity integer,
    is_unlimited boolean default false,
    daily_limit integer, -- Maximum orders per day
    
    -- Dietary features
    is_vegetarian boolean default false,
    is_vegan boolean default false,
    is_gluten_free boolean default false,
    is_halal boolean default false,
    is_spicy boolean default false,
    spice_level integer, -- 0-5 scale
    
    -- Nutritional values
    calories integer,
    protein decimal(5, 2),
    carbs decimal(5, 2),
    fat decimal(5, 2),
    
    -- Media
    image_url varchar(500),
    images text, -- JSON array: ["url1", "url2"]
    
    -- Statistics
    rating decimal(3, 2) default 0.00,
    total_reviews integer default 0,
    total_orders integer default 0,
    views integer default 0,
    
    -- Status
    is_active boolean default true,
    is_featured boolean default false,
    is_available boolean default true,
    
    -- Timestamps
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

-- Indexes
create index if not exists idx_menu_items_menu_id on menu_items(menu_id);
create index if not exists idx_menu_items_chef_id on menu_items(chef_id);
create index if not exists idx_menu_items_category on menu_items(category);
create index if not exists idx_menu_items_cuisine on menu_items(cuisine);
create index if not exists idx_menu_items_is_active on menu_items(is_active);
create index if not exists idx_menu_items_is_featured on menu_items(is_featured);
create index if not exists idx_menu_items_rating on menu_items(rating);
create index if not exists idx_menu_items_price on menu_items(price);
create index if not exists idx_menu_items_is_available on menu_items(is_available);

-- Comments
comment on table menu_items is 'Individual meals/dishes in menus';
comment on column menu_items.spice_level is '0=not spicy, 5=very spicy';
comment on column menu_items.is_unlimited is 'If true, available_quantity is ignored';
comment on column menu_items.is_available is 'Current availability status';
