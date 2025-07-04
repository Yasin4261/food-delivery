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
			return nil, nil // Sepet bulunamadı
		}
		return nil, err
	}
	
	return cart, nil
}

func (r *CartRepository) CreateCartItem(item *model.CartItem) error {
	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, item.CartID, item.ProductID, item.Quantity, 
		now, now).Scan(&item.ID)
	
	if err != nil {
		return err
	}
	
	item.CreatedAt = now
	item.UpdatedAt = now
	return nil
}

func (r *CartRepository) GetCartItem(cartID uint, productID uint) (*model.CartItem, error) {
	item := &model.CartItem{}
	query := `
		SELECT id, cart_id, product_id, quantity, created_at, updated_at
		FROM cart_items 
		WHERE cart_id = $1 AND product_id = $2`
	
	err := r.db.QueryRow(query, cartID, productID).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity,
		&item.CreatedAt, &item.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Item bulunamadı
		}
		return nil, err
	}
	
	return item, nil
}

func (r *CartRepository) GetCartItems(cartID uint) ([]model.CartItem, error) {
	query := `
		SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
		       p.name as product_name, p.price as product_price, p.image_url as product_image
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at DESC`
	
	rows, err := r.db.Query(query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem
		err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.Quantity,
			&item.CreatedAt, &item.UpdatedAt, &item.ProductName, 
			&item.ProductPrice, &item.ProductImage,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, nil
}

func (r *CartRepository) UpdateCartItem(item *model.CartItem) error {
	query := `
		UPDATE cart_items 
		SET quantity = $1, updated_at = $2
		WHERE id = $3`
	
	item.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, item.Quantity, item.UpdatedAt, item.ID)
	return err
}

func (r *CartRepository) DeleteCartItem(itemID uint) error {
	query := `DELETE FROM cart_items WHERE id = $1`
	_, err := r.db.Exec(query, itemID)
	return err
}

func (r *CartRepository) ClearCart(cartID uint) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err := r.db.Exec(query, cartID)
	return err
}

func (r *CartRepository) Delete(id uint) error {
	// Önce cart items'ları sil
	_, err := r.db.Exec(`DELETE FROM cart_items WHERE cart_id = $1`, id)
	if err != nil {
		return err
	}
	
	// Sonra cart'ı sil
	_, err = r.db.Exec(`DELETE FROM carts WHERE id = $1`, id)
	return err
}
