create table if not exists payment_methods (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,

    -- iyzico card wallet key (one per user) and the specific stored-card token.
    -- These are opaque gateway references; NO raw PAN/CVC is ever stored.
    card_user_key varchar(255) not null,
    card_token    varchar(255) not null,

    -- Display-only metadata returned by the gateway (masked digits + scheme).
    masked_number varchar(32)  not null,
    association   varchar(32),
    family        varchar(64),
    bank_name     varchar(128),

    created_at timestamp default current_timestamp,

    -- A stored card is unique per (user, gateway token); saving the same card
    -- twice is idempotent.
    unique (user_id, card_token)
);

create index if not exists idx_payment_methods_user_id on payment_methods(user_id);

comment on table payment_methods is 'Customer saved cards: iyzico cardUserKey/cardToken references + masked metadata only (no PAN/CVC)';
