package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateRepository_GetTemplatesByUserId(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewTemplateRepository(tx)

		// empty
		tmpls, err := repo.GetTemplatesByUserID(ctx, 999)
		require.NoError(t, err)
		assert.Empty(t, tmpls)

		// seed templates for user 1
		now := time.Now().UTC().Truncate(time.Second)
		_, err = tx.Exec(ctx,
			`INSERT INTO users(email, password_hash, created_at) VALUES ($1,$2,$3)`,
			"u@example.com", "hash", now,
		)
		require.NoError(t, err)

		// get inserted user id
		var userID int
		err = tx.QueryRow(ctx,
			`SELECT id FROM users WHERE email=$1`,
			"u@example.com",
		).Scan(&userID)
		require.NoError(t, err)

		// insert message_templates
		_, err = tx.Exec(ctx,
			`INSERT INTO message_templates(user_id, name, body, created_at, updated_at)
			 VALUES($1,$2,$3,$4,$4),($1,$5,$6,$4,$4)`,
			userID, "name1", "One", now, "name2", "Two",
		)
		require.NoError(t, err)

		// fetch
		tmpls, err = repo.GetTemplatesByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, tmpls, 2)
		bodies := []string{tmpls[0].Body, tmpls[1].Body}
		assert.Contains(t, bodies, "One")
		assert.Contains(t, bodies, "Two")
	})
}

func TestTemplateRepository_GetTemplateById(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewTemplateRepository(tx)

		// not exists
		tmpl, err := repo.GetTemplateByID(ctx, 1, 12345)
		assert.ErrorIs(t, err, domain.ErrTemplateNotExists)
		assert.Nil(t, tmpl)

		// seed user
		now := time.Now().UTC().Truncate(time.Second)
		var userID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash, created_at) VALUES($1,$2,$3) RETURNING id`,
			"b@example.com", "hash", now,
		).Scan(&userID)
		require.NoError(t, err)

		// seed template
		var tmplID int
		err = tx.QueryRow(ctx,
			`INSERT INTO message_templates(user_id, name, body, created_at, updated_at)
			 VALUES($1,$2,$3,$4,$4) RETURNING id`,
			userID, "Name", "Hello", now,
		).Scan(&tmplID)
		require.NoError(t, err)

		// fetch
		tmpl, err = repo.GetTemplateByID(ctx, userID, tmplID)
		require.NoError(t, err)
		assert.Equal(t, tmplID, tmpl.ID)
		assert.Equal(t, userID, tmpl.UserID)
		assert.Equal(t, "Hello", tmpl.Body)
		assert.WithinDuration(t, now, tmpl.CreationTime.UTC(), time.Second)
	})
}

func TestTemplateRepository_CreateTemplate(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewTemplateRepository(tx)

		// seed user
		var userID int
		err := tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash) VALUES($1,$2) RETURNING id`,
			"c@example.com", "hash",
		).Scan(&userID)
		require.NoError(t, err)

		// create
		tmpl := &models.Template{UserID: userID, Body: "NewBody"}
		created, err := repo.CreateTemplate(ctx, tmpl)
		require.NoError(t, err)
		assert.NotZero(t, created.ID)
		assert.Equal(t, userID, created.UserID)
		assert.Equal(t, "NewBody", created.Body)
	})
}

func TestTemplateRepository_UpdateTemplate(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewTemplateRepository(tx)

		// not exists
		_, err := repo.UpdateTemplate(ctx, 1, 9999, &models.Template{UserID: 1, Body: "X"})
		assert.ErrorIs(t, err, domain.ErrTemplateNotExists)

		// seed user
		var userID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash) VALUES($1,$2) RETURNING id`,
			"d@example.com", "hash",
		).Scan(&userID)
		require.NoError(t, err)

		// seed template
		var tmplID int
		err = tx.QueryRow(ctx,
			`INSERT INTO message_templates(user_id, name, body) VALUES($1,$2,$3) RETURNING id`,
			userID, "Old", "Old",
		).Scan(&tmplID)
		require.NoError(t, err)

		// update
		updated, err := repo.UpdateTemplate(ctx, userID, tmplID, &models.Template{UserID: userID, Body: "New"})
		require.NoError(t, err)
		assert.Equal(t, tmplID, updated.ID)
		assert.Equal(t, "New", updated.Body)
	})
}

func TestTemplateRepository_DeleteTemplate(t *testing.T) {
	testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
		repo := repository.NewTemplateRepository(tx)

		// not exists
		err := repo.DeleteTemplate(ctx, 1, 9999)
		assert.ErrorIs(t, err, domain.ErrTemplateNotExists)

		var userID, tmplID int
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash) VALUES($1,$2) RETURNING id`,
			"e@example.com", "hash",
		).Scan(&userID)
		require.NoError(t, err)
		err = tx.QueryRow(ctx,
			`INSERT INTO message_templates(user_id, name, body) VALUES($1,$2,$3) RETURNING id`,
			userID, "ToDel", "ToDel",
		).Scan(&tmplID)
		require.NoError(t, err)

		// delete
		err = repo.DeleteTemplate(ctx, userID, tmplID)
		require.NoError(t, err)
	})
}
