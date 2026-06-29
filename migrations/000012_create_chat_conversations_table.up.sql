create table if not exists chat_conversations (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    chef_id integer not null references chefs(id) on delete cascade,
    order_id integer references orders(id) on delete set null,

    last_message_at timestamp,
    created_at timestamp default current_timestamp,

    -- One thread per (customer, chef) pair.
    unique (user_id, chef_id)
);

create index if not exists idx_chat_conversations_user_id on chat_conversations(user_id);
create index if not exists idx_chat_conversations_chef_id on chat_conversations(chef_id);

comment on table chat_conversations is 'A message thread between a customer (user_id) and a chef (chef_id)';
