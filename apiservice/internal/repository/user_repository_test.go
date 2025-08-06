//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/repository"
	"github.com/stretchr/testify/require"
)

func clearUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func TestUserRepository_CRUD(t *testing.T) {
	t.Cleanup(func() { clearUsers(t, testDB) })

	ctx := context.Background()
	repo := repository.NewUserRepository(testPool)

	t.Run("CreateUser and GetUserByEmail/GetUserByID", func(t *testing.T) {
		orig := &models.User{Email: "foo@example.com", PasswordHash: "hashedpw"}
		created, err := repo.CreateUser(ctx, orig)
		require.NoError(t, err)
		require.NotZero(t, created.ID)
		require.Equal(t, orig.Email, created.Email)
		require.False(t, created.CreationTime.IsZero())

		byEmail, err := repo.GetUserByEmail(ctx, orig.Email)
		require.NoError(t, err)
		require.Equal(t, created.ID, byEmail.ID)
		require.Equal(t, orig.Email, byEmail.Email)

		byID, err := repo.GetUserByID(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.Email, byID.Email)
	})

	t.Run("GetUserByEmail NotExists", func(t *testing.T) {
		_, err := repo.GetUserByEmail(ctx, "no-such@example.com")
		require.ErrorIs(t, err, domain.ErrUserNotExists)
	})

	t.Run("GetUserByID NotExists", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, 9999)
		require.ErrorIs(t, err, domain.ErrUserNotExists)
	})

	t.Run("CreateUser DuplicateEmail", func(t *testing.T) {
		_, err := repo.CreateUser(ctx, &models.User{Email: "dup@example.com", PasswordHash: "h"})
		require.NoError(t, err)
		_, err = repo.CreateUser(ctx, &models.User{Email: "dup@example.com", PasswordHash: "h2"})
		require.Error(t, err)
	})
}
