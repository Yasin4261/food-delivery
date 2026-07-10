create table if not exists payment_sessions (
    id serial primary key,
    order_id integer not null references orders(id) on delete cascade,

    -- Gateway checkout token (from checkout-form initialize); single-use.
    token varchar(255) not null unique,
    -- Gateway payment id, set once the payment is verified paid.
    payment_id varchar(255),
    status varchar(20) not null default 'initiated', -- initiated, paid, failed, refunded

    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

create index if not exists idx_payment_sessions_order_id on payment_sessions(order_id);
create index if not exists idx_payment_sessions_token on payment_sessions(token);

comment on table payment_sessions is 'Hosted-checkout attempts per order (card payments via the PaymentGateway port)';
