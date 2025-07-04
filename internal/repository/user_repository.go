package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"time"
)

// UserRepository - kullanıcı veritabanı işlemleri
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (email, password, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, user.Email, user.Password, user.FirstName, 
		user.LastName, user.Role, now, now).Scan(&user.ID)
	
	if err != nil {
		return err
	}
	
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		FROM users 
		WHERE email = $1`
	
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Kullanıcı bulunamadı
		}
		return nil, err
	}
	
	return user, nil
}

func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		FROM users 
		WHERE id = $1`
	
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Kullanıcı bulunamadı
		}
		return nil, err
	}
	
	return user, nil
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	query := `
		SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.FirstName,
			&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

func (r *UserRepository) Update(user *model.User) error {
	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, role = $3, updated_at = $4
		WHERE id = $5`
	
	user.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, user.FirstName, user.LastName, user.Role, user.UpdatedAt, user.ID)
	return err
}

func (r *UserRepository) Delete(id uint) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) GetByRole(role string) ([]model.User, error) {
	query := `
		SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		FROM users 
		WHERE role = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.FirstName,
			&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}
