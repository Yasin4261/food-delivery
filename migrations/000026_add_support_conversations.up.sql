-- Support messaging (#120): a conversation can now be a customer<->chef thread
-- (kind='chef', the original shape) or an admin<->user support thread
-- (kind='support', chef_id NULL). This extends the existing chat_conversations
-- so support threads inherit the whole live-chat stack (WS hub, read receipts,
-- unread counts) rather than duplicating it.

-- 1. Discriminator. Existing rows are all customer<->chef threads.
alter table chat_conversations
    add column if not exists kind varchar(16) not null default 'chef';

-- 2. Support threads have no kitchen, so chef_id must be nullable.
alter table chat_conversations
    alter column chef_id drop not null;

-- 3. The old blanket unique(user_id, chef_id) can't express "one support thread
--    per user" (chef_id is NULL there). Replace it with two partial uniques:
--    chef threads stay one-per-(user,chef); support threads are one-per-user.
alter table chat_conversations
    drop constraint if exists chat_conversations_user_id_chef_id_key;

create unique index if not exists uq_chat_conversations_chef
    on chat_conversations(user_id, chef_id) where kind = 'chef';

create unique index if not exists uq_chat_conversations_support
    on chat_conversations(user_id) where kind = 'support';

-- 4. Integrity: a chef thread must name a kitchen, a support thread must not.
alter table chat_conversations
    add constraint chat_conversations_kind_shape check (
        (kind = 'chef' and chef_id is not null) or
        (kind = 'support' and chef_id is null)
    );

comment on column chat_conversations.kind is 'chef = customer<->chef thread; support = admin<->user thread (#120)';
