//go:build integration
// +build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/repository"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestGetNotificationByID(t *testing.T) {
	ctx := context.Background()

	loader := makeFixtures(t, testDB, "./../../../db/fixtures/notifications.yml")
	if err := loader.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	repo := repository.NewNotificationTasksRepository(testPool)

	existingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	nonExistID := uuid.New()

	tcs := []struct {
		name     string
		id       uuid.UUID
		wantErr  error
		assertFn func(*testing.T, *models.Notification)
	}{
		{
			name:    "found",
			id:      existingID,
			wantErr: nil,
			assertFn: func(t *testing.T, n *models.Notification) {
				assert.Equal(t, existingID, n.ID)
				assert.Equal(t, 1, n.UserID)
				assert.Equal(t, "+10000000001", n.RecipientPhone)
				assert.Equal(t, "pending", n.Status)
				assert.Equal(t, 0, n.Attempts)
			},
		},
		{
			name:     "not found",
			id:       nonExistID,
			wantErr:  domain.ErrNotificationNotExists,
			assertFn: func(t *testing.T, n *models.Notification) { assert.Nil(t, n) },
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			n, err := repo.GetNotificationByID(ctx, tc.id)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
			tc.assertFn(t, n)
		})
	}
}

func TestReschedule(t *testing.T) {
	ctx := context.Background()

	loader := makeFixtures(t, testDB, "./../../../db/fixtures/notifications.yml")
	if err := loader.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	repo := repository.NewNotificationTasksRepository(testPool)
	existingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	nonExistID := uuid.New()

	newRun := time.Now().Add(15 * time.Minute).Truncate(time.Second)

	tcs := []struct {
		name     string
		id       uuid.UUID
		nextRun  time.Time
		wantErr  error
		verifyDB bool
	}{
		{
			name:     "reschedule existing",
			id:       existingID,
			nextRun:  newRun,
			wantErr:  nil,
			verifyDB: true,
		},
		{
			name:     "reschedule non-existent",
			id:       nonExistID,
			nextRun:  newRun,
			wantErr:  domain.ErrNotificationNotExists,
			verifyDB: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			n, err := repo.Reschedule(ctx, tc.id, tc.nextRun)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "pending", n.Status)
			assert.WithinDuration(t, tc.nextRun, n.NextRunAt, time.Second)

			if tc.verifyDB {
				var rr time.Time
				err := testDB.QueryRowContext(ctx,
					`SELECT next_run_at FROM notifications WHERE id=$1`, tc.id,
				).Scan(&rr)
				assert.NoError(t, err)
				assert.WithinDuration(t, tc.nextRun, rr, time.Second)
			}
		})
	}
}

func TestMarkFailed(t *testing.T) {
	ctx := context.Background()

	loader := makeFixtures(t, testDB, "./../../../db/fixtures/notifications.yml")
	if err := loader.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	repo := repository.NewNotificationTasksRepository(testPool)
	existingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	nonExistID := uuid.New()

	tcs := []struct {
		name     string
		id       uuid.UUID
		wantErr  error
		verifyDB bool
	}{
		{
			name:     "mark existing failed",
			id:       existingID,
			wantErr:  nil,
			verifyDB: true,
		},
		{
			name:     "mark non-existent failed",
			id:       nonExistID,
			wantErr:  domain.ErrNotificationNotExists,
			verifyDB: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			n, err := repo.MarkFailed(ctx, tc.id)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "failed", n.Status)

			if tc.verifyDB {
				var st string
				err := testDB.QueryRowContext(ctx,
					`SELECT status FROM notifications WHERE id=$1`, tc.id,
				).Scan(&st)
				assert.NoError(t, err)
				assert.Equal(t, "failed", st)
			}
		})
	}
}
