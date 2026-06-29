create table if not exists chat_messages (
    id serial primary key,
    conversation_id integer not null references chat_conversations(id) on delete cascade,
    sender_id integer not null references users(id) on delete cascade,

    body text not null,
    read_at timestamp,
    created_at timestamp default current_timestamp
);

create index if not exists idx_chat_messages_conversation_id on chat_messages(conversation_id);
create index if not exists idx_chat_messages_created_at on chat_messages(created_at);

comment on table chat_messages is 'Messages within a chat conversation';
