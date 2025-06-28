package handler_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/testutils"
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

func mockUserMiddleware(userID any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), contextkeys.UserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
