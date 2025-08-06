//go:build integration
// +build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFetchAndUpdatePending(t *testing.T) {
	ctx := context.Background()

	loader := makeFixtures(t, testDB, "./../../../db/fixtures/notifications.yml")
	if err := loader.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	var dbConn domain.DBConn = testPool
	repo := repository.NewNotificationRepository(dbConn)

	t.Run("only pending and stale in-flight up to limit", func(t *testing.T) {
		const limit = 5
		notifs, err := repo.FetchAndUpdatePending(ctx, limit)
		if err != nil {
			t.Fatalf("FetchAndUpdatePending returned error: %v", err)
		}

		assert.Len(t, notifs, 2, "should return 2 notifications")

		got := make(map[uuid.UUID]*models.Notification, len(notifs))
		for _, n := range notifs {
			got[n.ID] = n
		}

		id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		n1, ok := got[id1]
		if !assert.True(t, ok, "expected id %s", id1) {
			t.FailNow()
		}
		assert.Equal(t, "in_flight", n1.Status)
		assert.Equal(t, 1, n1.Attempts)
		assert.WithinDuration(t, time.Now(), n1.UpdatedAt, time.Minute)

		id3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		n3, ok := got[id3]
		if !assert.True(t, ok, "expected id %s", id3) {
			t.FailNow()
		}
		assert.Equal(t, "in_flight", n3.Status)
		assert.Equal(t, 2, n3.Attempts)

		for _, id := range []uuid.UUID{id1, id3} {
			var status string
			var attempts int
			row := testDB.QueryRowContext(ctx,
				`SELECT status, attempts FROM notifications WHERE id = $1`, id,
			)
			if err := row.Scan(&status, &attempts); err != nil {
				t.Fatalf("failed scanning db for id=%s: %v", id, err)
			}
			assert.Equal(t, "in_flight", status)
			assert.Equal(t, got[id].Attempts, attempts)
		}
	})

	t.Run("limit smaller than available", func(t *testing.T) {
		if err := loader.Load(); err != nil {
			t.Fatalf("reload fixtures: %v", err)
		}
		const limit = 1
		notifs, err := repo.FetchAndUpdatePending(ctx, limit)
		if err != nil {
			t.Fatalf("FetchAndUpdatePending returned error: %v", err)
		}
		assert.Len(t, notifs, 1, "when limit=1 should fetch exactly one")
	})
}
