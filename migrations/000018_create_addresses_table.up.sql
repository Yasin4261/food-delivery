-- Customer address book (#66): saved delivery addresses with one default per
-- user. Orders snapshot the address text at placement, so rows here can be
-- edited or deleted freely without touching order history.
create table if not exists addresses (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,

    label varchar(50) not null,          -- "Home", "Work", ...
    address text not null,
    city varchar(100),
    latitude decimal(10, 8),
    longitude decimal(11, 8),
    is_default boolean not null default false,

    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);

create index if not exists idx_addresses_user_id on addresses(user_id);

-- At most one default address per user, enforced by the database.
create unique index if not exists idx_addresses_one_default
    on addresses(user_id) where is_default;
