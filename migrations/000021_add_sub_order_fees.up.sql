-- Money model (#65): per-slice delivery fee (distance-based, charged to the
-- customer) and platform commission (deducted from the chef's earnings).
-- Both are SNAPSHOTS taken at placement — rate changes never rewrite the
-- economics of historical orders.
alter table sub_orders
    add column if not exists delivery_fee decimal(10, 2) not null default 0.00,
    add column if not exists commission decimal(10, 2) not null default 0.00;
