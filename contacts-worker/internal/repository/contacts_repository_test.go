//go:build integration
// +build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestContactsRepository_SaveContacts(t *testing.T) {
	ctx := context.Background()

	repo := repository.NewContactsRepository(testPool)

	t.Run("inserts new contacts and ignores duplicates", func(t *testing.T) {
		fixtures := makeFixtures(t, testDB, "../../../db/fixtures/users.yml")
		if err := fixtures.Load(); err != nil {
			t.Fatalf("cannot load fixtures: %v", err)
		}

		contacts := []*models.Contact{
			{UserID: 1, Name: "Existing Contact", Phone: "123456789"}, // duplicate
			{UserID: 1, Name: "New Contact A", Phone: "999999999"},    // new
			{UserID: 2, Name: "New Contact B", Phone: "111111111"},    // new for another user
		}

		err := repo.SaveContacts(ctx, contacts)
		assert.NoError(t, err)

		rows, err := testPool.Query(ctx, `
			SELECT user_id, name, phone FROM contacts 
			ORDER BY user_id, phone`)
		assert.NoError(t, err)

		defer rows.Close()

		var results []models.Contact
		for rows.Next() {
			var c models.Contact
			err := rows.Scan(&c.UserID, &c.Name, &c.Phone)
			assert.NoError(t, err)
			results = append(results, c)
		}

		assert.Len(t, results, 3)
		assert.Contains(t, results, models.Contact{UserID: 1, Name: "Existing Contact", Phone: "123456789"})
		assert.Contains(t, results, models.Contact{UserID: 1, Name: "New Contact A", Phone: "999999999"})
		assert.Contains(t, results, models.Contact{UserID: 2, Name: "New Contact B", Phone: "111111111"})
	})
}
