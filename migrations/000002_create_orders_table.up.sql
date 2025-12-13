create table if not exists orders (
    id serial primary key,
    
    -- Order information
    order_code varchar(50) unique not null,
    user_id integer not null references users(id) on delete cascade,
    
    -- Pricing
    subtotal decimal(10, 2) not null,
    delivery_fee decimal(10, 2) default 0.00,
    service_fee decimal(10, 2) default 0.00,
    tax decimal(10, 2) default 0.00,
    discount decimal(10, 2) default 0.00,
    total_price decimal(10, 2) not null,
    
    -- Status and payment
    status varchar(20) not null default 'pending',
    payment_method varchar(20),
    payment_status varchar(20) default 'pending',
    
    -- Delivery information
    delivery_address text not null,
    delivery_city varchar(100),
    delivery_latitude decimal(10, 8),
    delivery_longitude decimal(11, 8),
    estimated_delivery_time timestamp,
    actual_delivery_time timestamp,
    
    -- Notes
    customer_notes text,
    chef_notes text,
    delivery_notes text,
    
    -- Timestamps
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,
    cancelled_at timestamp
);

-- Indexes
create index if not exists idx_orders_order_code on orders(order_code);
create index if not exists idx_orders_user_id on orders(user_id);
create index if not exists idx_orders_status on orders(status);
create index if not exists idx_orders_payment_status on orders(payment_status);
create index if not exists idx_orders_created_at on orders(created_at);

-- Comments
comment on table orders is 'Customer orders';
comment on column orders.status is 'Status: pending, confirmed, preparing, ready, delivering, delivered, cancelled';
comment on column orders.payment_status is 'Payment: pending, paid, failed, refunded';