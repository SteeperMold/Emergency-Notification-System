package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/testutils"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupProfileServerWithMockMiddleware(tx domain.DBConn, userID any) *mux.Router {
	logger := zap.NewNop()
	timeout := 2 * time.Second

	r := mux.NewRouter()

	r.Use(mockUserMiddleware(userID))

	route.NewProfileRoute(r, tx, logger, timeout)

	return r
}

func TestProfileHandler_GetProfile(t *testing.T) {
	type testCase struct {
		name          string
		setup         func(ctx context.Context, tx domain.DBConn) any
		userID        any
		expectedCode  int
		expectProfile bool
	}

	tests := []testCase{
		{
			name: "Success",
			setup: func(ctx context.Context, tx domain.DBConn) any {
				var id int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id`,
					"bob@example.com", "hashedpass",
				).Scan(&id)
				require.NoError(t, err)
				return id
			},
			expectedCode:  http.StatusOK,
			expectProfile: true,
		},
		{
			name:         "UserIdNotInt",
			userID:       "not-an-int",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "UserNotFound",
			userID:       999999,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
				userID := tc.userID
				if tc.setup != nil {
					userID = tc.setup(ctx, tx)
				}

				router := setupProfileServerWithMockMiddleware(tx, userID)
				srv := httptest.NewServer(router)
				defer srv.Close()

				req, err := http.NewRequest(http.MethodGet, srv.URL+"/profile", nil)
				require.NoError(t, err)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						panic(err)
					}
				}(resp.Body)

				require.Equal(t, tc.expectedCode, resp.StatusCode)

				if tc.expectProfile {
					var profile models.User
					err := json.NewDecoder(resp.Body).Decode(&profile)
					require.NoError(t, err)
					assert.Equal(t, userID, profile.ID)
				}
			})
		})
	}
}
