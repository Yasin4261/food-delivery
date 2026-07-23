-- Admin audit log (#121): every admin mutation records exactly one row here, in
-- the SAME transaction as the change it describes, so a rollback loses both and
-- no unattributed admin change can ever exist. Read-only (no update/delete).
create table if not exists admin_audit_log (
    id serial primary key,

    -- The admin who performed the action. NOT cascade-deleted: the trail must
    -- outlive the account (deletion anonymises rather than removes anyway).
    actor_user_id integer not null references users(id),

    action      varchar(64) not null,   -- e.g. 'chef.set_active'
    target_type varchar(32) not null,   -- 'user' | 'chef' | 'order' | 'promo'
    target_id   integer     not null,

    reason      text,                    -- required for destructive actions
    before_json jsonb,                   -- prior state (null for creates)
    after_json  jsonb,                   -- new state  (null for deletes)

    created_at timestamp default current_timestamp
);

create index if not exists idx_admin_audit_target on admin_audit_log(target_type, target_id);
create index if not exists idx_admin_audit_created_at on admin_audit_log(created_at desc);
create index if not exists idx_admin_audit_actor on admin_audit_log(actor_user_id);

comment on table admin_audit_log is 'Immutable trail of admin mutations, written atomically with each change (#121)';
