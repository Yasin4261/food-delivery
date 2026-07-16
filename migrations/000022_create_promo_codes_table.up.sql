-- Promo codes (#94): platform-funded discounts applied at checkout. A code is
-- either a percentage or a fixed amount off the food subtotal; the chef's
-- earnings are unaffected (the platform absorbs it). Limits: validity window,
-- minimum order amount, and a total usage cap enforced atomically.
create table if not exists promo_codes (
    id serial primary key,
    code varchar(40) not null unique,

    discount_type varchar(10) not null check (discount_type in ('percent', 'fixed')),
    discount_value decimal(10, 2) not null check (discount_value > 0),
    min_order decimal(10, 2) not null default 0,

    valid_from timestamp,
    valid_until timestamp,

    usage_limit integer not null default 0, -- 0 = unlimited
    used_count integer not null default 0,

    is_active boolean not null default true,
    created_at timestamp not null default current_timestamp
);

-- The applied code is snapshotted onto the order alongside orders.discount.
alter table orders add column if not exists promo_code varchar(40);
