create table if not exists favorites (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    chef_id integer not null references chefs(id) on delete cascade,

    created_at timestamp default current_timestamp,

    -- A customer can favorite a given chef at most once.
    unique (user_id, chef_id)
);

-- Indexes
create index if not exists idx_favorites_user_id on favorites(user_id);
create index if not exists idx_favorites_chef_id on favorites(chef_id);

-- Comments
comment on table favorites is 'Customers favoriting chefs';
