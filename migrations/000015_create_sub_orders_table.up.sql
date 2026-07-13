-- One row per (order, chef): the chef-scoped slice of a multi-chef order with
-- its own status lifecycle. The parent orders.status is derived from these
-- (all cancelled -> cancelled, all active delivered -> delivered, else the
-- least-advanced active sub-order). A sub-order's items are the order_items
-- rows matching (order_id, chef_id) — order_items is left untouched.
create table if not exists sub_orders (
    id serial primary key,
    order_id integer not null references orders(id) on delete cascade,
    chef_id integer not null references chefs(id),

    status varchar(20) not null default 'pending',
    subtotal decimal(10, 2) not null default 0.00,

    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,

    unique (order_id, chef_id)
);

create index if not exists idx_sub_orders_order_id on sub_orders(order_id);
create index if not exists idx_sub_orders_chef_id on sub_orders(chef_id);

-- Backfill: existing orders predate per-chef status, so every chef slice
-- inherits the parent order's status — history stays consistent.
insert into sub_orders (order_id, chef_id, status, subtotal, created_at, updated_at)
select oi.order_id, oi.chef_id, o.status, sum(oi.subtotal), o.created_at, o.updated_at
from order_items oi
join orders o on o.id = oi.order_id
group by oi.order_id, oi.chef_id, o.status, o.created_at, o.updated_at
on conflict (order_id, chef_id) do nothing;
