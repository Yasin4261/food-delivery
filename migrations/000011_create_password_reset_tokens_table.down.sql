drop index if exists idx_password_reset_tokens_user_id;
drop index if exists idx_password_reset_tokens_token_hash;
drop table if exists password_reset_tokens;
