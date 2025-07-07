package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"time"
)

// CartRepository - sepet veritabanı işlemleri
type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// Cart işlemleri
func (r *CartRepository) GetOrCreateCart(userID uint) (*model.Cart, error) {
	cart := &model.Cart{}
	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts 
		WHERE user_id = $1`
	
	err := r.db.QueryRow(query, userID).Scan(
		&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Sepet yoksa oluştur
			return r.CreateCart(userID)
		}
		return nil, err
	}
	
	return cart, nil
}

func (r *CartRepository) CreateCart(userID uint) (*model.Cart, error) {
	cart := &model.Cart{
		UserID: userID,
	}
	
	query := `
		INSERT INTO carts (user_id, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, cart.UserID, now, now).Scan(&cart.ID)
	
	if err != nil {
		return nil, err
	}
	
	cart.CreatedAt = now
	cart.UpdatedAt = now
	return cart, nil
}

// Eski metod adları ile uyumluluk
func (r *CartRepository) Create(cart *model.Cart) error {
	query := `
		INSERT INTO carts (user_id, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, cart.UserID, now, now).Scan(&cart.ID)
	
	if err != nil {
		return err
	}
	
	cart.CreatedAt = now
	cart.UpdatedAt = now
	return nil
}

func (r *CartRepository) GetByUserID(userID uint) (*model.Cart, error) {
	cart := &model.Cart{}
	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts 
		WHERE user_id = $1`
	
	err := r.db.QueryRow(query, userID).Scan(
		&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return cart, nil
}

// CartItem işlemleri
func (r *CartRepository) AddItem(cartID uint, mealID uint, chefID uint, quantity int, specialInstructions string) error {
	// Önce aynı yemek var mı kontrol et
	existingItem, err := r.GetItemByMealID(cartID, mealID)
	if err != nil {
		return err
	}
	
	if existingItem != nil {
		// Varsa miktarı güncelle
		return r.UpdateItemQuantity(existingItem.ID, existingItem.Quantity+quantity)
	}
	
	// Yoksa yeni item ekle
	query := `
		INSERT INTO cart_items (cart_id, meal_id, chef_id, quantity, special_instructions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	now := time.Now()
	_, err = r.db.Exec(query, cartID, mealID, chefID, quantity, specialInstructions, now, now)
	return err
}

func (r *CartRepository) GetItemByMealID(cartID uint, mealID uint) (*model.CartItem, error) {
	item := &model.CartItem{}
	query := `
		SELECT id, cart_id, meal_id, chef_id, quantity, special_instructions, created_at, updated_at
		FROM cart_items 
		WHERE cart_id = $1 AND meal_id = $2`
	
	err := r.db.QueryRow(query, cartID, mealID).Scan(
		&item.ID, &item.CartID, &item.MealID, &item.ChefID, &item.Quantity,
		&item.SpecialInstructions, &item.CreatedAt, &item.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return item, nil
}

func (r *CartRepository) UpdateItemQuantity(itemID uint, quantity int) error {
	if quantity <= 0 {
		return r.RemoveItem(itemID)
	}
	
	query := `
		UPDATE cart_items 
		SET quantity = $1, updated_at = $2
		WHERE id = $3`
	
	_, err := r.db.Exec(query, quantity, time.Now(), itemID)
	return err
}

func (r *CartRepository) RemoveItem(itemID uint) error {
	query := `DELETE FROM cart_items WHERE id = $1`
	_, err := r.db.Exec(query, itemID)
	return err
}

func (r *CartRepository) GetCartItems(cartID uint) ([]model.CartItem, error) {
	query := `
		SELECT id, cart_id, meal_id, chef_id, quantity, special_instructions, created_at, updated_at
		FROM cart_items 
		WHERE cart_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem
		err := rows.Scan(
			&item.ID, &item.CartID, &item.MealID, &item.ChefID, &item.Quantity,
			&item.SpecialInstructions, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, nil
}

func (r *CartRepository) GetCartItemsWithMeals(cartID uint) ([]model.CartItemResponse, error) {
	query := `
		SELECT ci.id, ci.meal_id, ci.chef_id, ci.quantity, ci.special_instructions,
			ci.created_at, ci.updated_at,
			m.name as meal_name, m.price as meal_price, m.images as meal_image,
			CONCAT(u.first_name, ' ', u.last_name) as chef_name, c.kitchen_name
		FROM cart_items ci
		JOIN meals m ON ci.meal_id = m.id
		JOIN chefs c ON ci.chef_id = c.id
		JOIN users u ON c.user_id = u.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at DESC`
	
	rows, err := r.db.Query(query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []model.CartItemResponse
	for rows.Next() {
		var item model.CartItemResponse
		err := rows.Scan(
			&item.ID, &item.MealID, &item.ChefID, &item.Quantity, &item.SpecialInstructions,
			&item.CreatedAt, &item.UpdatedAt,
			&item.MealName, &item.MealPrice, &item.MealImage,
			&item.ChefName, &item.KitchenName,
		)
		if err != nil {
			return nil, err
		}
		
		item.Subtotal = item.MealPrice * float64(item.Quantity)
		items = append(items, item)
	}
	
	return items, nil
}

func (r *CartRepository) ClearCart(cartID uint) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err := r.db.Exec(query, cartID)
	return err
}

// GetCartItem - sepet öğesini al (cart service uyumluluk için)
func (r *CartRepository) GetCartItem(cartID uint, mealID uint) (*model.CartItem, error) {
	return r.GetItemByMealID(cartID, mealID)
}

// UpdateCartItem - sepet öğesini güncelle (cart service uyumluluk için)
func (r *CartRepository) UpdateCartItem(item *model.CartItem) error {
	query := `
		UPDATE cart_items 
		SET quantity = $1, special_instructions = $2, updated_at = $3
		WHERE id = $4`
	
	_, err := r.db.Exec(query, item.Quantity, item.SpecialInstructions, time.Now(), item.ID)
	return err
}

// CreateCartItem - sepet öğesi oluştur (cart service uyumluluk için)
func (r *CartRepository) CreateCartItem(item *model.CartItem) error {
	query := `
		INSERT INTO cart_items (cart_id, meal_id, chef_id, quantity, special_instructions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, item.CartID, item.MealID, item.ChefID, item.Quantity, 
		item.SpecialInstructions, now, now).Scan(&item.ID)
	
	if err != nil {
		return err
	}
	
	item.CreatedAt = now
	item.UpdatedAt = now
	return nil
}

// DeleteCartItem - sepet öğesi sil (cart service uyumluluk için)
func (r *CartRepository) DeleteCartItem(itemID uint) error {
	return r.RemoveItem(itemID)
}
