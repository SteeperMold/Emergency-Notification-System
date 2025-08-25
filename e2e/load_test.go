//go:build load
// +build load

package e2e

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/bxcodec/faker/v4"
	"github.com/stretchr/testify/require"
)

// Contact is the schema for generating mock contacts
type Contact struct {
	Name  string `faker:"name"`
	Phone string `faker:"russian_phone"`
}

// custom provider for Russian-style phone numbers
func russianPhoneNumber(v reflect.Value) (interface{}, error) {
	num := "+79"
	for i := 0; i < 9; i++ {
		num += fmt.Sprintf("%d", rand.Intn(10))
	}
	return num, nil
}

func generateMockContactsCSV(path string, n int) error {
	rand.Seed(time.Now().UnixNano())
	_ = faker.AddProvider("russian_phone", russianPhoneNumber)

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	_ = w.Write([]string{"name", "phone"})

	for i := 0; i < n; i++ {
		var c Contact
		if err := faker.FakeData(&c); err != nil {
			return err
		}
		_ = w.Write([]string{c.Name, c.Phone})
	}

	return nil
}

func TestMillionContactsNotification(t *testing.T) {
	const (
		csvPath   = "testdata/mock_contacts.csv"
		nContacts = 1_000_000
	)

	// generate CSV if missing
	_, err := os.Stat(csvPath)
	if os.IsNotExist(err) {
		t.Logf("generating %d contacts into %s...", nContacts, csvPath)
		require.NoError(t, generateMockContactsCSV(csvPath, nContacts))
	} else {
		t.Logf("using existing contacts file: %s", csvPath)
	}

	t.Log("counting sent notifications before the test...")
	initialNtfsCount := FetchSentNotificationsCount(t)

	t.Log("signing up...")
	accessToken := SendSignupRequest(t, "test@e2eload.com", "123456789admin")
	t.Log("posting template...")
	templateID := PostTemplateRequest(t, accessToken, "Load test template name", "Load test template body")

	t.Log("counting contacts before the test...")
	initialContactsCount := GetContactsRequest(t, accessToken)
	t.Log("loading contacts from file...")
	SendLoadContactsFromFileRequest(t, accessToken, csvPath)

	loadContactsDeadline := time.Now().Add(60 * time.Second)
	contactsLoaded := false
	var expectedNtfsCount int

	for time.Now().Before(loadContactsDeadline) {
		contactsCount := GetContactsRequest(t, accessToken)
		newContacts := contactsCount - initialContactsCount
		t.Logf("waiting for contacts processing; current count: %v", newContacts)

		if float64(newContacts) >= float64(nContacts)*0.99 {
			contactsLoaded = true
			expectedNtfsCount = newContacts
			break
		}

		time.Sleep(1 * time.Second)
	}

	require.True(t, contactsLoaded)
	t.Logf("successfully loaded contacts; sending notifications...")

	SendNotificationRequest(t, accessToken, templateID)

	sendNtfsDeadline := time.Now().Add(30 * time.Minute)
	var sentNtfsCount int

	for time.Now().Before(sendNtfsDeadline) {
		sentNtfsCount = FetchSentNotificationsCount(t) - initialNtfsCount
		t.Logf("waiting for notifications sending; current count: %v", sentNtfsCount)

		if float64(sentNtfsCount) >= float64(expectedNtfsCount)*0.9999 {
			break
		}
		time.Sleep(10 * time.Second)
	}

	successRate := float64(sentNtfsCount) / float64(expectedNtfsCount)
	require.GreaterOrEqual(t, successRate, 0.9999, "too little success rate: %v", successRate)

	t.Log("successfully delivered 99.99%+ of notifications")
}
