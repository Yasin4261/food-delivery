-- Migration: Multi-chef (multi-vendor) order support
-- Bu migration, tek bir siparişin birden fazla şeften ürün içerebilmesi için gerekli yapıları ekler

-- Sub-orders tablosu: Ana sipariş içindeki şef bazlı alt siparişler
CREATE TABLE IF NOT EXISTS sub_orders (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    chef_id INTEGER NOT NULL REFERENCES chefs(id) ON DELETE CASCADE,
    subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    chef_notes TEXT,
    estimated_preparation_time INTEGER, -- dakika cinsinden
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(order_id, chef_id) -- Bir siparişte aynı şeften sadece bir sub-order olabilir
);

-- Order tablosuna konum desteği ekle
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_latitude DECIMAL(10, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_longitude DECIMAL(11, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_type VARCHAR(20) DEFAULT 'delivery';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'cash';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS customer_note TEXT;

-- Order items tablosuna sub_order_id referansı ekle
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS sub_order_id INTEGER REFERENCES sub_orders(id) ON DELETE CASCADE;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS special_instructions TEXT;

-- Index'ler ekle (performans için)
CREATE INDEX IF NOT EXISTS idx_sub_orders_order_id ON sub_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_sub_orders_chef_id ON sub_orders(chef_id);
CREATE INDEX IF NOT EXISTS idx_sub_orders_status ON sub_orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_sub_order_id ON order_items(sub_order_id);
CREATE INDEX IF NOT EXISTS idx_orders_delivery_location ON orders(delivery_latitude, delivery_longitude);
CREATE INDEX IF NOT EXISTS idx_orders_delivery_type ON orders(delivery_type);

-- Chef tablosuna konum ve teslimat yarıçapı bilgisi ekle (konum bazlı öneriler için)
ALTER TABLE chefs ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8);
ALTER TABLE chefs ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);
ALTER TABLE chefs ADD COLUMN IF NOT EXISTS delivery_radius INTEGER DEFAULT 10; -- km cinsinden
ALTER TABLE chefs ADD COLUMN IF NOT EXISTS average_preparation_time INTEGER DEFAULT 30; -- dakika cinsinden
ALTER TABLE chefs ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT true;

-- Chef tablosuna index ekle
CREATE INDEX IF NOT EXISTS idx_chefs_location ON chefs(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_chefs_available ON chefs(is_available);

-- Meal tablosuna ek bilgiler
ALTER TABLE meals ADD COLUMN IF NOT EXISTS preparation_time INTEGER DEFAULT 20; -- dakika cinsinden
ALTER TABLE meals ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT true;

-- Meal tablosuna index ekle
CREATE INDEX IF NOT EXISTS idx_meals_available ON meals(is_available);

-- Konum bazlı chef bulma için fonksiyon (haversine formula)
CREATE OR REPLACE FUNCTION calculate_distance(lat1 DECIMAL, lon1 DECIMAL, lat2 DECIMAL, lon2 DECIMAL)
RETURNS DECIMAL AS $$
DECLARE
    earth_radius DECIMAL := 6371; -- km cinsinden Dünya yarıçapı
    lat1_rad DECIMAL := lat1 * PI() / 180;
    lat2_rad DECIMAL := lat2 * PI() / 180;
    delta_lat DECIMAL := (lat2 - lat1) * PI() / 180;
    delta_lon DECIMAL := (lon2 - lon1) * PI() / 180;
    a DECIMAL;
    c DECIMAL;
BEGIN
    a := SIN(delta_lat/2) * SIN(delta_lat/2) + 
         COS(lat1_rad) * COS(lat2_rad) * 
         SIN(delta_lon/2) * SIN(delta_lon/2);
    c := 2 * ATAN2(SQRT(a), SQRT(1-a));
    RETURN earth_radius * c;
END;
$$ LANGUAGE plpgsql;

-- Test verileri için bazı lokasyon bilgileri ekle
UPDATE chefs SET 
    latitude = 41.0082 + (RANDOM() - 0.5) * 0.1,  -- İstanbul merkez civarı
    longitude = 28.9784 + (RANDOM() - 0.5) * 0.1,
    delivery_radius = 5 + FLOOR(RANDOM() * 10),    -- 5-15 km arası
    average_preparation_time = 20 + FLOOR(RANDOM() * 40), -- 20-60 dakika arası
    is_available = true
WHERE latitude IS NULL;

-- Trigger: Sub-order oluşturulduğunda ana siparişin totalini güncelle
CREATE OR REPLACE FUNCTION update_order_total()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE orders 
    SET total = (
        SELECT COALESCE(SUM(subtotal), 0) 
        FROM sub_orders 
        WHERE order_id = NEW.order_id
    )
    WHERE id = NEW.order_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_order_total
    AFTER INSERT OR UPDATE OR DELETE ON sub_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_order_total();

-- Trigger: Order item oluşturulduğunda sub-order subtotal'ini güncelle  
CREATE OR REPLACE FUNCTION update_sub_order_subtotal()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.sub_order_id IS NOT NULL THEN
        UPDATE sub_orders 
        SET subtotal = (
            SELECT COALESCE(SUM(quantity * price), 0) 
            FROM order_items 
            WHERE sub_order_id = NEW.sub_order_id
        )
        WHERE id = NEW.sub_order_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_sub_order_subtotal
    AFTER INSERT OR UPDATE OR DELETE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_sub_order_subtotal();
