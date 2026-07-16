-- Email verification (#103): after registration a single-use, expiring token is
-- emailed to the account owner; redeeming it flips users.is_verified. Only the
-- sha256 hash of the raw token is stored, mirroring password_reset_tokens — the
-- raw token travels to the user out of band and is never persisted or returned.
create table if not exists email_verification_tokens (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    token_hash varchar(64) not null unique,
    expires_at timestamp not null,
    used_at timestamp,
    created_at timestamp not null default current_timestamp
);

create index if not exists idx_email_verification_tokens_user_id
    on email_verification_tokens (user_id);
