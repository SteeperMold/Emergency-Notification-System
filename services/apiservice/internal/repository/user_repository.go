package repository

import (
	"context"
	"errors"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/jackc/pgx/v5"
)

// UserRepository provides access to the user storage.
type UserRepository struct {
	db domain.DBConn
}

// NewUserRepository creates a new UserRepository with the given database connection.
func NewUserRepository(db domain.DBConn) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreateUser inserts a new user into the database and returns the created user with ID and timestamps.
func (ur *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`

	var newUser models.User

	row := ur.db.QueryRow(ctx, q, user.Email, user.PasswordHash)
	err := row.Scan(&newUser.ID, &newUser.Email, &newUser.CreationTime)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

// GetUserByEmail fetches a user by email. Returns domain.ErrUserNotExists if no user is found.
func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var user models.User

	row := ur.db.QueryRow(ctx, q, email)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreationTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotExists
		}

		return nil, err
	}

	return &user, nil
}

// GetUserByID fetches a user by ID. Returns domain.ErrUserNotExists if no user is found.
func (ur *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user models.User

	row := ur.db.QueryRow(ctx, q, id)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreationTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotExists
		}

		return nil, err
	}

	return &user, nil
}
