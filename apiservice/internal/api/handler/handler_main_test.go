package handler

import (
	"os"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/internal/testutils"
)

func TestMain(m *testing.M) {
	_, err := testutils.SetupTestDB()
	if err != nil {
		panic(err)
	}

	code := m.Run()

	testutils.TeardownTestDB()

	os.Exit(code)
}
