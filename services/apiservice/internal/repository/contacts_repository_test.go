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

func clearContacts(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE contacts RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func TestContactsRepository_GetAllContactsByUserID(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	require.NoError(t, fixtures.Load())

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	contact1 := &models.Contact{UserID: userID, Name: "Foo", Phone: "+111"}
	contact2 := &models.Contact{UserID: userID, Name: "Bar", Phone: "+222"}
	_, err := repo.CreateContact(ctx, contact1)
	require.NoError(t, err)
	_, err = repo.CreateContact(ctx, contact2)
	require.NoError(t, err)

	contacts, err := repo.GetAllContactsByUserID(ctx, userID)
	require.NoError(t, err)
	require.Len(t, contacts, 2)
	require.ElementsMatch(t, []string{"Foo", "Bar"}, []string{contacts[0].Name, contacts[1].Name})
}

func TestContactsRepository_GetContactsCountByUserID(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	require.NoError(t, fixtures.Load())

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	count, err := repo.GetContactsCountByUserID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	contact := &models.Contact{UserID: userID, Name: "Counted", Phone: "+999"}
	_, err = repo.CreateContact(ctx, contact)
	require.NoError(t, err)

	count, err = repo.GetContactsCountByUserID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestCreateAndGetContact(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1

	repo := repository.NewContactsRepository(testPool)

	t.Run("create contact", func(t *testing.T) {

		contact := &models.Contact{UserID: userID, Name: "John", Phone: "+123456789"}
		created, err := repo.CreateContact(ctx, contact)
		require.NoError(t, err)
		require.Equal(t, contact.Name, created.Name)
		require.Equal(t, contact.Phone, created.Phone)
	})

	t.Run("get contact by id", func(t *testing.T) {
		contact := &models.Contact{UserID: userID, Name: "Alice", Phone: "+198765432"}
		created, err := repo.CreateContact(ctx, contact)
		require.NoError(t, err)
		got, err := repo.GetContactByID(ctx, userID, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.Name, got.Name)
		require.Equal(t, created.Phone, got.Phone)
	})
}

func TestContactsRepository_GetContactsByUserID(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	contactA := &models.Contact{UserID: userID, Name: "Alpha", Phone: "+111"}
	contactB := &models.Contact{UserID: userID, Name: "Beta", Phone: "+222"}
	_, err := repo.CreateContact(ctx, contactA)
	require.NoError(t, err)
	_, err = repo.CreateContact(ctx, contactB)
	require.NoError(t, err)

	contacts, err := repo.GetContactsPageByUserID(ctx, userID, 100, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(contacts), 2)
	names := []string{contacts[0].Name, contacts[1].Name}
	require.Contains(t, names, "Alpha")
	require.Contains(t, names, "Beta")
}

func TestContactsRepository_GetContactByID_NotExists(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	repo := repository.NewContactsRepository(testPool)

	_, err := repo.GetContactByID(ctx, 9999, 9999)
	require.ErrorIs(t, err, domain.ErrContactNotExists)
}

func TestContactsRepository_CreateContact_UniqueConstraint(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	contact := &models.Contact{UserID: userID, Name: "Dup", Phone: "+999"}
	_, err := repo.CreateContact(ctx, contact)
	require.NoError(t, err)

	_, err = repo.CreateContact(ctx, contact)
	require.ErrorIs(t, err, domain.ErrContactAlreadyExists)
}

func TestContactsRepository_UpdateContact(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	orig := &models.Contact{UserID: userID, Name: "Old", Phone: "+333"}
	created, err := repo.CreateContact(ctx, orig)
	require.NoError(t, err)

	t.Run("successfully update", func(t *testing.T) {
		updatedInfo := &models.Contact{UserID: userID, Name: "NewName", Phone: "+444"}
		updated, err := repo.UpdateContact(ctx, userID, created.ID, updatedInfo)
		require.NoError(t, err)
		require.Equal(t, "NewName", updated.Name)
		require.Equal(t, "+444", updated.Phone)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.UpdateContact(ctx, userID, 0, orig)
		require.ErrorIs(t, err, domain.ErrContactNotExists)
	})

	t.Run("unique violation", func(t *testing.T) {
		// create a second contact
		other := &models.Contact{UserID: userID, Name: "Other", Phone: "+555"}
		_, err := repo.CreateContact(ctx, other)
		require.NoError(t, err)

		// attempt to update 'created' to have same name+phone as 'other'
		conflict := &models.Contact{UserID: userID, Name: other.Name, Phone: other.Phone}
		_, err = repo.UpdateContact(ctx, userID, created.ID, conflict)
		require.ErrorIs(t, err, domain.ErrContactAlreadyExists)
	})
}

func TestContactsRepository_DeleteContact(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	// insert and delete
	contact := &models.Contact{UserID: userID, Name: "ToDelete", Phone: "+666"}
	created, err := repo.CreateContact(ctx, contact)
	require.NoError(t, err)

	err = repo.DeleteContact(ctx, userID, created.ID)
	require.NoError(t, err)

	// deleting again should return not‐exists
	err = repo.DeleteContact(ctx, userID, created.ID)
	require.ErrorIs(t, err, domain.ErrContactNotExists)
}

func TestContactsRepository_DeleteContact_WrongUser(t *testing.T) {
	t.Cleanup(func() { clearContacts(t, testDB) })

	fixtures := makeFixtures(t, testDB, "../../../../db/fixtures/users.yml")
	if err := fixtures.Load(); err != nil {
		t.Fatalf("failed loading fixtures: %v", err)
	}

	ctx := context.Background()
	userID := 1
	repo := repository.NewContactsRepository(testPool)

	contact := &models.Contact{UserID: userID, Name: "WrongUser", Phone: "+777"}
	created, err := repo.CreateContact(ctx, contact)
	require.NoError(t, err)

	// attempt delete as different user
	err = repo.DeleteContact(ctx, userID+1, created.ID)
	require.ErrorIs(t, err, domain.ErrContactNotExists)
}
