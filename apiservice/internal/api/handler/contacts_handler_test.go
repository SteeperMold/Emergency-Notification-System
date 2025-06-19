package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/testutils"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContactsServer(tx domain.DBConn, userID any) *mux.Router {
	logger := zap.NewNop()
	timeout := 2 * time.Second

	r := mux.NewRouter()
	if userID != nil {
		r.Use(mockUserMiddleware(userID))
	}
	route.NewContactsRoute(r, tx, logger, timeout) // you’ll need to add NewContactsRoute in your route package
	return r
}

type contactsTC struct {
	name           string
	method         string
	path           string
	body           interface{}
	useTx          bool
	setup          func(ctx context.Context, tx domain.DBConn) (any, string)
	expectedStatus int
	expectContains string
	expectJSON     bool
}

func TestContactsHandler(t *testing.T) {
	cases := []contactsTC{
		// --- GET /contacts --------------------------------
		{
			name:           "Get_NoUserID",
			method:         http.MethodGet,
			path:           "/contacts",
			useTx:          false,
			expectedStatus: http.StatusInternalServerError,
			expectContains: "internal server error",
		},
		{
			name:   "Get_EmptyList",
			method: http.MethodGet,
			path:   "/contacts",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				// seed one user, no contacts
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"a@a.com", "h",
				).Scan(&uid))
				return uid, ""
			},
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},

		// --- GET /contacts/{id} --------------------------
		{
			name:           "GetByID_BadID",
			method:         http.MethodGet,
			path:           "/contacts/abc",
			useTx:          true,
			setup:          func(ctx context.Context, tx domain.DBConn) (any, string) { return 1, "" },
			expectedStatus: http.StatusBadRequest,
			expectContains: "invalid id",
		},
		{
			name:   "GetByID_NotFound",
			method: http.MethodGet,
			path:   "/contacts/9999",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"u@u.com", "h",
				).Scan(&uid))
				return uid, ""
			},
			expectedStatus: http.StatusNotFound,
			expectContains: "contact not exists",
		},
		{
			name:   "GetByID_Success",
			method: http.MethodGet,
			path:   "/contacts/",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, cid int
				now := time.Now().UTC().Truncate(time.Second)
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash,created_at) VALUES($1,$2,$3) RETURNING id`,
					"x@x.com", "h", now,
				).Scan(&uid))
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO contacts(user_id,name,phone,created_at,updated_at) VALUES($1,$2,$3,$4,$4) RETURNING id`,
					uid, "N", "+1000000000", now,
				).Scan(&cid))
				return uid, strconv.Itoa(cid)
			},
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},

		// --- POST /contacts -----------------------------
		{
			name:   "Post_InvalidJSON",
			method: http.MethodPost,
			path:   "/contacts",
			body:   "not a json",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"p@p.com", "h",
				).Scan(&uid))
				return uid, ""
			},
			expectedStatus: http.StatusBadRequest,
			expectContains: "invalid request",
		},
		{
			name:   "Post_InvalidContact",
			method: http.MethodPost,
			path:   "/contacts",
			body:   map[string]string{"name": "", "phone": "123"}, // invalid
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"q@q.com", "h",
				).Scan(&uid))
				return uid, ""
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectContains: domain.ErrInvalidContact.Error(),
		},
		{
			name:   "Post_Success",
			method: http.MethodPost,
			path:   "/contacts",
			body:   map[string]string{"name": "Alice", "phone": "+7 912 456 78 90"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"r@r.com", "h",
				).Scan(&uid))
				return uid, ""
			},
			expectedStatus: http.StatusCreated,
			expectJSON:     true,
		},
		{
			name:   "Post_Conflict",
			method: http.MethodPost,
			path:   "/contacts",
			body:   map[string]string{"name": "Bob", "phone": "+7 912 456 78 90"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"s@s.com", "h",
				).Scan(&uid))
				// first insert
				_, err := tx.Exec(ctx,
					`INSERT INTO contacts(user_id,name,phone) VALUES($1,$2,$3)`,
					uid, "Bob", "+4 123 456 78 90",
				)
				require.NoError(t, err)
				return uid, ""
			},
			expectedStatus: http.StatusConflict,
			expectContains: "contact already exists",
		},

		// --- PUT /contacts/{id} ------------------------
		{
			name:           "Put_BadID",
			method:         http.MethodPut,
			path:           "/contacts/xyz",
			body:           map[string]string{"name": "X", "phone": "+7 912 456 78 90"},
			useTx:          true,
			setup:          func(ctx context.Context, tx domain.DBConn) (any, string) { return 1, "" },
			expectedStatus: http.StatusBadRequest,
			expectContains: "invalid id",
		},
		{
			name:   "Put_NotFound",
			method: http.MethodPut,
			path:   "/contacts/",
			body:   map[string]string{"name": "Y", "phone": "+7 912 456 78 90"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"t@t.com", "h",
				).Scan(&uid))
				return uid, "9999"
			},
			expectedStatus: http.StatusNotFound,
			expectContains: "contact not exists",
		},
		{
			name:   "Put_Success",
			method: http.MethodPut,
			path:   "/contacts/",
			body:   map[string]string{"name": "Zed", "phone": "+7 912 456 78 90"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, cid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"u2@u2.com", "h",
				).Scan(&uid))
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO contacts(user_id,name,phone) VALUES($1,$2,$3) RETURNING id`,
					uid, "Old", "+8 123 456 78 90",
				).Scan(&cid))
				return uid, strconv.Itoa(cid)
			},
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},

		// --- DELETE /contacts/{id} ---------------------
		{
			name:           "Delete_BadID",
			method:         http.MethodDelete,
			path:           "/contacts/xyz",
			useTx:          true,
			setup:          func(ctx context.Context, tx domain.DBConn) (any, string) { return 1, "" },
			expectedStatus: http.StatusBadRequest,
			expectContains: "invalid id",
		},
		{
			name:   "Delete_NotFound",
			method: http.MethodDelete,
			path:   "/contacts/",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"d2@d2.com", "h",
				).Scan(&uid))
				return uid, "9999"
			},
			expectedStatus: http.StatusNotFound,
			expectContains: "contact doesn't exist",
		},
		{
			name:   "Delete_Success",
			method: http.MethodDelete,
			path:   "/contacts/",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, cid int
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"f@f.com", "h",
				).Scan(&uid))
				require.NoError(t, tx.QueryRow(ctx,
					`INSERT INTO contacts(user_id,name,phone) VALUES($1,$2,$3) RETURNING id`,
					uid, "Fay", "+90000000000",
				).Scan(&cid))
				return uid, strconv.Itoa(cid)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func(ctx context.Context, tx domain.DBConn) {
				var userID any
				var idStr string
				if tc.setup != nil {
					userID, idStr = tc.setup(ctx, tx)
				}
				runContactsTest(t, tx, tc, userID, idStr)
			}
			if tc.useTx {
				testutils.WithRollback(t, run)
			} else {
				run(context.Background(), nil)
			}
		})
	}
}

func runContactsTest(
	t *testing.T,
	tx domain.DBConn,
	tc contactsTC,
	userID any,
	idStr string,
) {
	router := setupContactsServer(tx, userID)
	srv := httptest.NewServer(router)
	defer srv.Close()

	var bodyReader io.Reader
	if tc.body != nil {
		bs, _ := json.Marshal(tc.body)
		bodyReader = bytes.NewReader(bs)
	}

	url := srv.URL + tc.path
	if idStr != "" {
		url += idStr
	}

	req, err := http.NewRequest(tc.method, url, bodyReader)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	require.Equal(t, tc.expectedStatus, resp.StatusCode, string(bodyBytes))

	if tc.expectJSON {
		var out interface{}
		require.NoError(t, json.Unmarshal(bodyBytes, &out))
	} else if tc.expectContains != "" {
		assert.Contains(t, string(bodyBytes), tc.expectContains)
	}
}
