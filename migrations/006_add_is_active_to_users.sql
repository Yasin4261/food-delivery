-- +goose Up
-- Users tablosuna is_active sütunu ekle
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Var olan kullanıcılar için is_active değerini true yap
UPDATE users SET is_active = true WHERE is_active IS NULL;

-- +goose Down
-- Users tablosundan is_active sütununu kaldır
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
