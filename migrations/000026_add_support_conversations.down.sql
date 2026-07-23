-- Reverse of 000026. Support threads cannot survive the restored NOT NULL +
-- unique(user_id, chef_id), so they are dropped (dev-only rollback; this loses
-- support conversation data).
delete from chat_conversations where kind = 'support';

alter table chat_conversations drop constraint if exists chat_conversations_kind_shape;
drop index if exists uq_chat_conversations_support;
drop index if exists uq_chat_conversations_chef;

alter table chat_conversations alter column chef_id set not null;
alter table chat_conversations
    add constraint chat_conversations_user_id_chef_id_key unique (user_id, chef_id);

alter table chat_conversations drop column if exists kind;
