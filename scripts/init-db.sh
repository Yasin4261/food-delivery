#!/bin/bash
set -e

# Veritabanı oluştur
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Veritabanı zaten docker-compose ile oluşturuluyor
    -- Bu dosya ek konfigürasyonlar için kullanılabilir
    
    -- Örnek: Varsayılan veriler eklemek için
    -- INSERT INTO categories (name) VALUES ('Elektronik'), ('Giyim'), ('Kitap');
EOSQL

echo "Veritabanı başarıyla hazırlandı!"
