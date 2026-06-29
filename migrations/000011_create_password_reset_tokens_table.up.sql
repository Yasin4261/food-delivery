create table if not exists password_reset_tokens (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,

    -- Only a hash of the token is stored; the raw token is delivered to the user.
    token_hash varchar(64) not null unique,
    expires_at timestamp not null,
    used_at timestamp,

    created_at timestamp default current_timestamp
);

create index if not exists idx_password_reset_tokens_token_hash on password_reset_tokens(token_hash);
create index if not exists idx_password_reset_tokens_user_id on password_reset_tokens(user_id);

comment on table password_reset_tokens is 'Single-use, expiring password reset tokens (token_hash = sha256 of the raw token)';
