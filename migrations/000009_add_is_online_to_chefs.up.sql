-- Presence: whether the chef is currently online. Distinct from
-- is_accepting_orders ("open for business") — they answer different questions.
alter table chefs add column if not exists is_online boolean not null default false;

create index if not exists idx_chefs_is_online on chefs(is_online);

comment on column chefs.is_online is 'Live presence toggle, separate from is_accepting_orders';
