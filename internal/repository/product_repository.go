package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"time"
)

// ProductRepository - ürün veritabanı işlemleri
type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *model.Product) error {
	query := `
		INSERT INTO products (name, description, price, stock, category_id, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, product.Name, product.Description, product.Price, 
		product.Stock, product.CategoryID, product.ImageURL, now, now).Scan(&product.ID)
	
	if err != nil {
		return err
	}
	
	product.CreatedAt = now
	product.UpdatedAt = now
	return nil
}

func (r *ProductRepository) GetAll() ([]model.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, 
		       p.created_at, p.updated_at, c.name as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var products []model.Product
	for rows.Next() {
		var product model.Product
		var categoryName sql.NullString
		
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.CategoryID, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &categoryName,
		)
		if err != nil {
			return nil, err
		}
		
		if categoryName.Valid {
			product.CategoryName = categoryName.String
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	product := &model.Product{}
	query := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, 
		       p.created_at, p.updated_at, c.name as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1`
	
	var categoryName sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.CategoryID, &product.ImageURL,
		&product.CreatedAt, &product.UpdatedAt, &categoryName,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Ürün bulunamadı
		}
		return nil, err
	}
	
	if categoryName.Valid {
		product.CategoryName = categoryName.String
	}
	
	return product, nil
}

func (r *ProductRepository) GetByCategory(categoryID uint) ([]model.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, 
		       p.created_at, p.updated_at, c.name as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = $1
		ORDER BY p.created_at DESC`
	
	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var products []model.Product
	for rows.Next() {
		var product model.Product
		var categoryName sql.NullString
		
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.CategoryID, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &categoryName,
		)
		if err != nil {
			return nil, err
		}
		
		if categoryName.Valid {
			product.CategoryName = categoryName.String
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

func (r *ProductRepository) Search(query string) ([]model.Product, error) {
	searchQuery := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, 
		       p.created_at, p.updated_at, c.name as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.name ILIKE '%' || $1 || '%' OR p.description ILIKE '%' || $1 || '%'
		ORDER BY p.created_at DESC`
	
	rows, err := r.db.Query(searchQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var products []model.Product
	for rows.Next() {
		var product model.Product
		var categoryName sql.NullString
		
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.CategoryID, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &categoryName,
		)
		if err != nil {
			return nil, err
		}
		
		if categoryName.Valid {
			product.CategoryName = categoryName.String
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

func (r *ProductRepository) Update(product *model.Product) error {
	query := `
		UPDATE products 
		SET name = $1, description = $2, price = $3, stock = $4, 
		    category_id = $5, image_url = $6, updated_at = $7
		WHERE id = $8`
	
	product.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, product.Name, product.Description, product.Price,
		product.Stock, product.CategoryID, product.ImageURL, product.UpdatedAt, product.ID)
	return err
}

func (r *ProductRepository) Delete(id uint) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ProductRepository) GetLowStock(threshold int) ([]model.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, 
		       p.created_at, p.updated_at, c.name as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.stock <= $1
		ORDER BY p.stock ASC`
	
	rows, err := r.db.Query(query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var products []model.Product
	for rows.Next() {
		var product model.Product
		var categoryName sql.NullString
		
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.CategoryID, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &categoryName,
		)
		if err != nil {
			return nil, err
		}
		
		if categoryName.Valid {
			product.CategoryName = categoryName.String
		}
		
		products = append(products, product)
	}
	
	return products, nil
}
