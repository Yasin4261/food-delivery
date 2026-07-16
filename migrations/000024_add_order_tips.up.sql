-- Tips (#105): an optional gratuity the customer adds at checkout. It is added
-- to the order total (the customer pays it) and goes to the chef uncommissioned.
-- The order-level amount is snapshotted on orders.tip; for multi-chef carts it
-- is split proportionally by food subtotal and snapshotted per slice on
-- sub_orders.tip, which feeds chef earnings and decline refunds.
alter table orders add column if not exists tip decimal(10, 2) not null default 0;
alter table sub_orders add column if not exists tip decimal(10, 2) not null default 0;
