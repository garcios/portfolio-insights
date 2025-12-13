// Package repository implements data access for the user domain.
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/pkg/database"
	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/user-service/internal/metrics"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// GetByID retrieves a user by ID from the customers.users table
func (r *userRepository) GetByID(id string) (*domain.User, error) {
	start := time.Now()
	defer func() {
		database.RecordQuery("get_by_id", "users", time.Since(start).Seconds(), nil)
	}()

	query := `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM customers.users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		database.RecordQuery("get_by_id", "users", time.Since(start).Seconds(), err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Create inserts a new user into the customers.users table
func (r *userRepository) Create(user *domain.User) error {
	start := time.Now()
	query := `
		INSERT INTO customers.users (email, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		user.Email,
		user.Username,
		user.Password,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		database.RecordQuery("create", "users", time.Since(start).Seconds(), err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	database.RecordQuery("create", "users", time.Since(start).Seconds(), nil)
	database.RecordRowsAffected("create", "users", 1)
	metrics.RecordUserCreated()
	return nil
}

// GetByEmail retrieves a user by email from the customers.users table
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	start := time.Now()
	defer func() {
		database.RecordQuery("get_by_email", "users", time.Since(start).Seconds(), nil)
	}()

	query := `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM customers.users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		database.RecordQuery("get_by_email", "users", time.Since(start).Seconds(), err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user in the customers.users table
func (r *userRepository) Update(user *domain.User) error {
	start := time.Now()
	query := `
		UPDATE customers.users
		SET email = $1, username = $2, password_hash = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		user.Email,
		user.Username,
		user.Password,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err == sql.ErrNoRows {
		database.RecordQuery("update", "users", time.Since(start).Seconds(), domain.ErrUserNotFound)
		return domain.ErrUserNotFound
	}
	if err != nil {
		database.RecordQuery("update", "users", time.Since(start).Seconds(), err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	database.RecordQuery("update", "users", time.Since(start).Seconds(), nil)
	database.RecordRowsAffected("update", "users", 1)
	return nil
}

// Delete removes a user from the customers.users table
func (r *userRepository) Delete(id string) error {
	start := time.Now()
	query := `DELETE FROM customers.users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		database.RecordQuery("delete", "users", time.Since(start).Seconds(), err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		database.RecordQuery("delete", "users", time.Since(start).Seconds(), err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		database.RecordQuery("delete", "users", time.Since(start).Seconds(), domain.ErrUserNotFound)
		return domain.ErrUserNotFound
	}

	database.RecordQuery("delete", "users", time.Since(start).Seconds(), nil)
	database.RecordRowsAffected("delete", "users", rowsAffected)
	return nil
}

// Count returns the total number of users
func (r *userRepository) Count() (int, error) {
	start := time.Now()
	defer func() {
		database.RecordQuery("count", "users", time.Since(start).Seconds(), nil)
	}()

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM customers.users").Scan(&count)
	if err != nil {
		database.RecordQuery("count", "users", time.Since(start).Seconds(), err)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
