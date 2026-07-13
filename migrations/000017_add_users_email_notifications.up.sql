-- Per-user opt-out for order notification emails (#71). Defaults to on;
-- security email (password reset) is not governed by this flag.
alter table users
    add column if not exists email_notifications boolean not null default true;
