package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateUser(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewUserRepository(tx)

		created, err := repo.CreateUser(ctx, &models.User{
			Email:        "new@example.com",
			PasswordHash: "pwhash",
		})
		require.NoError(t, err)
		assert.Equal(t, "new@example.com", created.Email)
		assert.NotZero(t, created.ID)
		assert.WithinDuration(t, time.Now(), created.CreationTime.UTC(), time.Second)

		row := tx.QueryRow(ctx,
			`SELECT email, password_hash FROM users WHERE id = $1`,
			created.ID,
		)
		var email, hash string
		err = row.Scan(&email, &hash)
		require.NoError(t, err)
		assert.Equal(t, "new@example.com", email)
		assert.Equal(t, "pwhash", hash)

		_, err = repo.CreateUser(ctx, &models.User{
			Email:        "new@example.com",
			PasswordHash: "another",
		})
		assert.Error(t, err)
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewUserRepository(tx)

		user, err := repo.GetUserByEmail(ctx, "nope@example.com")
		assert.ErrorIs(t, err, domain.ErrUserNotExists)
		assert.Nil(t, user)

		now := time.Now().UTC().Truncate(time.Second)
		_, err = tx.Exec(ctx,
			`INSERT INTO users (email, password_hash, created_at) VALUES ($1,$2,$3)`,
			"foo@bar.com", "hashpw", now,
		)
		require.NoError(t, err)

		got, err := repo.GetUserByEmail(ctx, "foo@bar.com")
		require.NoError(t, err)
		assert.Equal(t, "foo@bar.com", got.Email)
		assert.Equal(t, "hashpw", got.PasswordHash)
		assert.WithinDuration(t, now, got.CreationTime.UTC(), time.Second)
	})
}

func TestUserRepository_GetUserById(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewUserRepository(tx)

		user, err := repo.GetUserByID(ctx, 9999)
		assert.ErrorIs(t, err, domain.ErrUserNotExists)
		assert.Nil(t, user)

		now := time.Now().UTC().Truncate(time.Second)
		_, err = tx.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, created_at) VALUES ($1,$2,$3,$4)`,
			1, "foo@bar.com", "hashpw", now,
		)
		require.NoError(t, err)

		got, err := repo.GetUserByID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "foo@bar.com", got.Email)
		assert.Equal(t, "hashpw", got.PasswordHash)
		assert.WithinDuration(t, now, got.CreationTime.UTC(), time.Second)
	})
}
