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

func TestContactsRepository_GetContactsByUserID(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewContactsRepository(tx)

		// empty
		cts, err := repo.GetContactsByUserID(ctx, 999)
		require.NoError(t, err)
		assert.Empty(t, cts)

		// seed a user
		now := time.Now().UTC().Truncate(time.Second)
		var userID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash, created_at) VALUES($1,$2,$3) RETURNING id`,
			"user@example.com", "hash", now,
		).Scan(&userID)
		require.NoError(t, err)

		// insert contacts
		_, err = tx.Exec(ctx,
			`INSERT INTO contacts(user_id, name, phone, created_at, updated_at)
			   VALUES ($1,$2,$3,$4,$4),($1,$5,$6,$4,$4)`,
			userID, "Alice", "+10000000000", now, "Bob", "+20000000000",
		)
		require.NoError(t, err)

		// fetch
		cts, err = repo.GetContactsByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, cts, 2)
		names := []string{cts[0].Name, cts[1].Name}
		assert.Contains(t, names, "Alice")
		assert.Contains(t, names, "Bob")
	})
}

func TestContactsRepository_GetContactByID(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewContactsRepository(tx)

		// not exists
		ct, err := repo.GetContactByID(ctx, 1, 12345)
		assert.ErrorIs(t, err, domain.ErrContactNotExists)
		assert.Nil(t, ct)

		// seed user
		now := time.Now().UTC().Truncate(time.Second)
		var userID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash, created_at) VALUES($1,$2,$3) RETURNING id`,
			"bob@example.com", "hash", now,
		).Scan(&userID)
		require.NoError(t, err)

		// seed contact
		var contactID int
		err = tx.QueryRow(ctx,
			`INSERT INTO contacts(user_id, name, phone, created_at, updated_at)
			   VALUES($1,$2,$3,$4,$4) RETURNING id`,
			userID, "Charlie", "+30000000000", now,
		).Scan(&contactID)
		require.NoError(t, err)

		// fetch
		ct, err = repo.GetContactByID(ctx, userID, contactID)
		require.NoError(t, err)
		assert.Equal(t, contactID, ct.ID)
		assert.Equal(t, userID, ct.UserID)
		assert.Equal(t, "Charlie", ct.Name)
		assert.Equal(t, "+30000000000", ct.Phone)
		assert.WithinDuration(t, now, ct.CreationTime.UTC(), time.Second)
	})
}

func TestContactsRepository_CreateContact(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewContactsRepository(tx)

		// seed user
		var userID int
		err := tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash) VALUES($1,$2) RETURNING id`,
			"carol@example.com", "hash",
		).Scan(&userID)
		require.NoError(t, err)

		// create
		ctr := &models.Contact{UserID: userID, Name: "Dave", Phone: "+40000000000"}
		created, err := repo.CreateContact(ctx, ctr)
		require.NoError(t, err)
		assert.NotZero(t, created.ID)
		assert.Equal(t, userID, created.UserID)
		assert.Equal(t, "Dave", created.Name)
		assert.Equal(t, "+40000000000", created.Phone)

		// duplicate -> already exists
		_, err = repo.CreateContact(ctx, ctr)
		assert.ErrorIs(t, err, domain.ErrContactAlreadyExists)
	})
}

func TestContactsRepository_UpdateContact(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewContactsRepository(tx)

		// not exists
		_, err := repo.UpdateContact(ctx, 1, 9999, &models.Contact{UserID: 1, Name: "X", Phone: "+50000000000"})
		assert.ErrorIs(t, err, domain.ErrContactNotExists)

		// seed user and contact
		var userID, contactID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash) VALUES($1,$2) RETURNING id`,
			"ed@example.com", "hash",
		).Scan(&userID)
		require.NoError(t, err)
		err = tx.QueryRow(ctx,
			`INSERT INTO contacts(user_id, name, phone) VALUES($1,$2,$3) RETURNING id`,
			userID, "Eve", "+60000000000",
		).Scan(&contactID)
		require.NoError(t, err)

		// update
		updated, err := repo.UpdateContact(ctx, userID, contactID, &models.Contact{UserID: userID, Name: "Eve2", Phone: "+70000000000"})
		require.NoError(t, err)
		assert.Equal(t, contactID, updated.ID)
		assert.Equal(t, "Eve2", updated.Name)
		assert.Equal(t, "+70000000000", updated.Phone)

		// conflict on unique (same name+phone)
		// seed another
		_, err = tx.Exec(ctx,
			`INSERT INTO contacts(user_id, name, phone) VALUES($1,$2,$3)`,
			userID, "Dup", "+80000000000",
		)
		require.NoError(t, err)
		// attempt to rename our record to collide
		_, err = repo.UpdateContact(ctx, userID, contactID, &models.Contact{UserID: userID, Name: "Dup", Phone: "+80000000000"})
		assert.ErrorIs(t, err, domain.ErrContactAlreadyExists)
	})
}

func TestContactsRepository_DeleteContact(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewContactsRepository(tx)

		// not exists
		err := repo.DeleteContact(ctx, 1, 9999)
		assert.ErrorIs(t, err, domain.ErrContactNotExists)

		// seed and delete
		var userID, contactID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash) VALUES($1,$2) RETURNING id`,
			"frank@example.com", "hash",
		).Scan(&userID)
		require.NoError(t, err)
		err = tx.QueryRow(ctx,
			`INSERT INTO contacts(user_id, name, phone) VALUES($1,$2,$3) RETURNING id`,
			userID, "Frank", "+90000000000",
		).Scan(&contactID)
		require.NoError(t, err)

		// delete
		err = repo.DeleteContact(ctx, userID, contactID)
		require.NoError(t, err)
	})
}
