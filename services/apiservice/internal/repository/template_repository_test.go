//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/repository"
	"github.com/stretchr/testify/require"
)

func clearTemplates(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE message_templates RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func TestTemplateRepository_CRUD(t *testing.T) {
	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1
	repo := repository.NewTemplateRepository(testPool)

	t.Run("Create and GetByID", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		input := &models.Template{UserID: userID, Name: "Hello", Body: "World"}
		created, err := repo.CreateTemplate(ctx, input)
		require.NoError(t, err)
		require.Equal(t, input.Name, created.Name)
		require.Equal(t, input.Body, created.Body)

		fetched, err := repo.GetTemplateByID(ctx, userID, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, fetched.ID)
	})

	t.Run("List Templates", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		_, err := repo.CreateTemplate(ctx, &models.Template{UserID: userID, Name: "T1", Body: "B1"})
		require.NoError(t, err)
		_, err = repo.CreateTemplate(ctx, &models.Template{UserID: userID, Name: "T2", Body: "B2"})
		require.NoError(t, err)

		list, err := repo.GetTemplatesByUserID(ctx, userID, 100, 0)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(list), 2)
		names := []string{list[0].Name, list[1].Name}
		require.Contains(t, names, "T1")
		require.Contains(t, names, "T2")
	})

	t.Run("GetByID_NotExists", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		_, err := repo.GetTemplateByID(ctx, userID, 9999)
		require.ErrorIs(t, err, domain.ErrTemplateNotExists)
	})

	t.Run("Create_UniqueViolation", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		base := &models.Template{UserID: userID, Name: "Dup", Body: "Same"}
		_, err := repo.CreateTemplate(ctx, base)
		require.NoError(t, err)

		_, err = repo.CreateTemplate(ctx, base)
		require.ErrorIs(t, err, domain.ErrTemplateAlreadyExists)
	})

	t.Run("UpdateTemplate", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		orig, err := repo.CreateTemplate(ctx, &models.Template{UserID: userID, Name: "Old", Body: "B"})
		require.NoError(t, err)

		t.Run("success", func(t *testing.T) {
			upd := &models.Template{UserID: userID, Name: "Old", Body: "NewB"}
			updated, err := repo.UpdateTemplate(ctx, userID, orig.ID, upd)
			require.NoError(t, err)
			require.Equal(t, "NewB", updated.Body)
		})

		t.Run("not found", func(t *testing.T) {
			_, err := repo.UpdateTemplate(ctx, userID, 0, &models.Template{UserID: userID, Name: "X", Body: "Y"})
			require.ErrorIs(t, err, domain.ErrTemplateNotExists)
		})

		t.Run("unique violation", func(t *testing.T) {
			other, err := repo.CreateTemplate(ctx, &models.Template{UserID: userID, Name: "Other", Body: "Z"})
			require.NoError(t, err)

			conflict := &models.Template{UserID: userID, Name: other.Name, Body: other.Body}
			_, err = repo.UpdateTemplate(ctx, userID, orig.ID, conflict)
			require.ErrorIs(t, err, domain.ErrTemplateAlreadyExists)
		})
	})

	t.Run("DeleteTemplate", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		toDel, err := repo.CreateTemplate(ctx, &models.Template{UserID: userID, Name: "D", Body: "B"})
		require.NoError(t, err)

		err = repo.DeleteTemplate(ctx, userID, toDel.ID)
		require.NoError(t, err)

		err = repo.DeleteTemplate(ctx, userID, toDel.ID)
		require.ErrorIs(t, err, domain.ErrTemplateNotExists)
	})

	t.Run("DeleteTemplate_WrongUser", func(t *testing.T) {
		t.Cleanup(func() { clearTemplates(t, testDB) })

		tmpl, err := repo.CreateTemplate(ctx, &models.Template{UserID: userID, Name: "W", Body: "B"})
		require.NoError(t, err)

		err = repo.DeleteTemplate(ctx, userID+1, tmpl.ID)
		require.ErrorIs(t, err, domain.ErrTemplateNotExists)
	})
}
