create table if not exists reviews (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    order_id integer not null references orders(id) on delete cascade,

    -- A review targets exactly one of: a chef, or a dish (menu item).
    chef_id integer references chefs(id) on delete cascade,
    menu_item_id integer references menu_items(id) on delete cascade,

    rating integer not null check (rating between 1 and 5),
    comment text,

    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,

    -- Exactly one target must be set.
    constraint reviews_one_target check (
        (chef_id is not null and menu_item_id is null)
        or (chef_id is null and menu_item_id is not null)
    ),

    -- A customer may review a given chef/dish only once per order.
    unique (user_id, order_id, chef_id),
    unique (user_id, order_id, menu_item_id)
);

-- Indexes
create index if not exists idx_reviews_chef_id on reviews(chef_id);
create index if not exists idx_reviews_menu_item_id on reviews(menu_item_id);
create index if not exists idx_reviews_user_id on reviews(user_id);
create index if not exists idx_reviews_order_id on reviews(order_id);

-- Comments
comment on table reviews is 'Customer ratings of chefs and dishes; feeds chefs.rating and menu_items.rating';
