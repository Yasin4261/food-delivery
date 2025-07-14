-- Test verisi ekleyelim

-- Admin kullanıcı
INSERT INTO users (email, password, first_name, last_name, role) VALUES
('admin@example.com', '$2a$10$hash', 'Admin', 'User', 'admin'),
('ahmet@example.com', '$2a$10$hash', 'Ahmet', 'Yılmaz', 'chef'),
('ayse@example.com', '$2a$10$hash', 'Ayşe', 'Kaya', 'chef'),
('mehmet@example.com', '$2a$10$hash', 'Mehmet', 'Demir', 'customer'),
('fatma@example.com', '$2a$10$hash', 'Fatma', 'Özkan', 'customer');

-- Chef'ler
INSERT INTO chefs (user_id, business_name, description, location, is_active, is_verified, rating, total_orders) VALUES
(2, 'Ahmet Usta Mutfağı', 'Geleneksel Türk mutfağı ustası', 'İstanbul, Kadıköy', true, true, 4.5, 25),
(3, 'Ayşe Hanım Ev Yemekleri', 'Ev yapımı sağlıklı yemekler', 'Ankara, Çankaya', true, true, 4.8, 42);

-- Yemekler
INSERT INTO meals (chef_id, name, description, price, category, ingredients, allergens, prep_time, portion_size, is_available) VALUES
(1, 'Ev Yapımı Karnıyarık', 'Taze patlıcan ve kıyma ile hazırlanmış geleneksel karnıyarık', 35.00, 'Ana Yemek', 'Patlıcan, kıyma, domates, soğan, baharat', 'Yok', 45, 1, true),
(1, 'Mercimek Çorbası', 'Evde hazırlanmış nefis mercimek çorbası', 18.00, 'Çorba', 'Mercimek, havuç, soğan, baharat', 'Yok', 30, 1, true),
(2, 'Ev Yapımı Sarma', 'Taze asma yaprağı ile hazırlanmış sarma', 28.00, 'Ana Yemek', 'Asma yaprağı, pirinç, kıyma, baharatlar', 'Yok', 60, 1, true),
(2, 'Tavuk Sote', 'Sebzeli tavuk sote', 32.00, 'Ana Yemek', 'Tavuk, biber, domates, soğan', 'Yok', 25, 1, true),
(1, 'Baklava', 'Ev yapımı baklava', 25.00, 'Tatlı', 'Yufka, ceviz, şerbet', 'Gluten, Ceviz', 120, 1, true);

-- Sepetler
INSERT INTO carts (user_id) VALUES (4), (5);

-- Sepet öğeleri
INSERT INTO cart_items (cart_id, meal_id, chef_id, quantity) VALUES
(1, 1, 1, 2),
(1, 2, 1, 1),
(2, 3, 2, 1),
(2, 4, 2, 1);

-- Siparişler
INSERT INTO orders (user_id, chef_id, total, status, address, delivery_date, delivery_time) VALUES
(4, 1, 88.00, 'completed', 'Ataşehir, İstanbul', '2024-01-15', '19:00:00'),
(5, 2, 60.00, 'pending', 'Bahçelievler, Ankara', '2024-01-16', '20:00:00');

-- Sipariş öğeleri
INSERT INTO order_items (order_id, meal_id, chef_id, quantity, price) VALUES
(1, 1, 1, 2, 35.00),
(1, 2, 1, 1, 18.00),
(2, 3, 2, 1, 28.00),
(2, 4, 2, 1, 32.00);

-- Değerlendirmeler
INSERT INTO reviews (user_id, meal_id, chef_id, order_id, rating, comment) VALUES
(4, 1, 1, 1, 5, 'Çok lezzetliydi, evde yemek yemiş gibi hissettim!'),
(4, 2, 1, 1, 4, 'Çorba çok güzeldi, biraz daha tuzlu olabilirdi.');
