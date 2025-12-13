create table if not exists order_items (
    id serial primary key,
    order_id integer not null references orders(id) on delete cascade,
    menu_item_id integer not null references menu_items(id),
    chef_id integer not null references chefs(id),
    
    -- Item information (snapshot - prices can change)
    item_name varchar(200) not null,
    quantity integer not null default 1,
    unit_price decimal(10, 2) not null,
    subtotal decimal(10, 2) not null,
    
    -- Special requests
    special_instructions text,
    
    -- Timestamps
    created_at timestamp default current_timestamp
);

-- Indexes
create index if not exists idx_order_items_order_id on order_items(order_id);
create index if not exists idx_order_items_menu_item_id on order_items(menu_item_id);
create index if not exists idx_order_items_chef_id on order_items(chef_id);

-- Comments
comment on table order_items is 'Individual items within orders';
comment on column order_items.item_name is 'Snapshot of item name at order time';
comment on column order_items.subtotal is 'quantity * unit_price';
