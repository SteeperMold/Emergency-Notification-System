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

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/testutils"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupSignupServer(tx domain.DBConn) *mux.Router {
	logger := zap.NewNop()
	timeout := 2 * time.Second
	jwtConfig := &bootstrap.JWTConfig{
		AccessSecret:  "access",
		AccessExpiry:  2 * time.Hour,
		RefreshSecret: "refresh",
		RefreshExpiry: 720 * time.Hour,
	}

	r := mux.NewRouter()
	route.NewSignupRouter(r, tx, logger, timeout, jwtConfig)

	return r
}

func TestSignupHandler(t *testing.T) {
	type testCase struct {
		name           string
		setup          func(ctx context.Context, tx domain.DBConn)
		useTx          bool
		request        domain.SignupRequest
		expectedStatus int
		expectBody     string
		expectEmail    string
		expectUserID   bool
	}

	testCases := []testCase{
		{
			name:  "Success",
			useTx: true,
			setup: nil,
			request: domain.SignupRequest{
				Email:    "bob@example.com",
				Password: "strongpass",
			},
			expectedStatus: http.StatusCreated,
			expectEmail:    "bob@example.com",
			expectUserID:   true,
		},
		{
			name:  "InvalidEmail",
			useTx: false,
			request: domain.SignupRequest{
				Email:    "bad",
				Password: "strongpass",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectBody:     "invalid email",
		},
		{
			name:  "InvalidPassword",
			useTx: false,
			request: domain.SignupRequest{
				Email:    "alice@example.com",
				Password: "bad",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectBody:     "invalid password",
		},
		{
			name:  "EmailAlreadyExists",
			useTx: true,
			setup: func(ctx context.Context, tx domain.DBConn) {
				_, err := tx.Exec(ctx,
					`INSERT INTO users(email, password_hash) VALUES($1, $2)`,
					"alice@example.com", "hashpw",
				)
				require.NoError(t, err)
			},
			request: domain.SignupRequest{
				Email:    "alice@example.com",
				Password: "anotherpass",
			},
			expectedStatus: http.StatusConflict,
			expectBody:     "email already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.useTx {
				testutils.WithRollback(t, func(ctx context.Context, tx domain.DBConn) {
					if tc.setup != nil {
						tc.setup(ctx, tx)
					}
					runSignupTest(t, tx, tc)
				})
			} else {
				runSignupTest(t, nil, tc)
			}
		})
	}
}

func runSignupTest(t *testing.T, tx domain.DBConn, tc struct {
	name           string
	setup          func(ctx context.Context, tx domain.DBConn)
	useTx          bool
	request        domain.SignupRequest
	expectedStatus int
	expectBody     string
	expectEmail    string
	expectUserID   bool
}) {
	router := setupSignupServer(tx)
	srv := httptest.NewServer(router)
	defer srv.Close()

	reqBody, _ := json.Marshal(tc.request)
	resp, err := http.Post(srv.URL+"/signup", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	assert.Equal(t, tc.expectedStatus, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)

	if tc.expectedStatus == http.StatusCreated {
		var u domain.SignupResponse
		err = json.Unmarshal(body, &u)
		require.NoError(t, err)
		assert.Equal(t, tc.expectEmail, u.User.Email)
		if tc.expectUserID {
			assert.NotZero(t, u.User.ID)
		}
	} else {
		assert.Contains(t, string(body), tc.expectBody)
	}
}
