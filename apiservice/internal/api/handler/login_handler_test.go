package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/testutils"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func setupLoginServer(tx domain.DBConn) *mux.Router {
	logger := zap.NewNop()
	timeout := 2 * time.Second
	jwtConfig := &bootstrap.JWTConfig{
		AccessSecret:  "access",
		AccessExpiry:  2 * time.Hour,
		RefreshSecret: "refresh",
		RefreshExpiry: 720 * time.Hour,
	}

	r := mux.NewRouter()
	route.NewLoginRoute(r, tx, logger, timeout, jwtConfig)

	return r
}

func TestLoginHandler(t *testing.T) {
	type testCase struct {
		name           string
		setupUser      func(ctx context.Context, tx domain.DBConn)
		email          string
		password       string
		expectedStatus int
		expectBody     string
		expectEmail    string
		expectTokens   bool
	}

	testCases := []testCase{
		{
			name: "Success",
			setupUser: func(ctx context.Context, tx domain.DBConn) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("strongpass"), bcrypt.DefaultCost)
				_, err := tx.Exec(ctx,
					`INSERT INTO users(email, password_hash) VALUES($1, $2)`,
					"bob@example.com", string(hash),
				)
				require.NoError(t, err)
			},
			email:          "bob@example.com",
			password:       "strongpass",
			expectedStatus: http.StatusOK,
			expectEmail:    "bob@example.com",
			expectTokens:   true,
		},
		{
			name:           "InvalidEmail",
			setupUser:      nil,
			email:          "bad",
			password:       "strongpass",
			expectedStatus: http.StatusUnauthorized,
			expectBody:     "invalid credentials",
		},
		{
			name: "InvalidPassword",
			setupUser: func(ctx context.Context, tx domain.DBConn) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
				_, err := tx.Exec(ctx,
					`INSERT INTO users(email, password_hash) VALUES($1, $2)`,
					"alice@example.com", string(hash),
				)
				require.NoError(t, err)
			},
			email:          "alice@example.com",
			password:       "wrongpass",
			expectedStatus: http.StatusUnauthorized,
			expectBody:     "invalid credentials",
		},
		{
			name:           "UserNotFound",
			setupUser:      nil,
			email:          "nonexistent@example.com",
			password:       "somepass",
			expectedStatus: http.StatusUnauthorized,
			expectBody:     "invalid credentials",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
				if tc.setupUser != nil {
					tc.setupUser(ctx, tx)
				}

				router := setupLoginServer(tx)
				srv := httptest.NewServer(router)
				defer srv.Close()

				reqBody, _ := json.Marshal(domain.LoginRequest{
					Email:    tc.email,
					Password: tc.password,
				})

				resp, err := http.Post(srv.URL+"/login", "application/json", bytes.NewReader(reqBody))
				require.NoError(t, err)
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						panic(err)
					}
				}(resp.Body)

				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				body, _ := io.ReadAll(resp.Body)

				if tc.expectedStatus == http.StatusOK {
					var lr domain.LoginResponse
					err = json.Unmarshal(body, &lr)
					require.NoError(t, err)

					assert.Equal(t, tc.expectEmail, lr.User.Email)

					if tc.expectTokens {
						assert.NotEmpty(t, lr.AccessToken)
						assert.NotEmpty(t, lr.RefreshToken)
					}
				} else {
					assert.Contains(t, string(body), tc.expectBody)
				}
			})
		})
	}
}
