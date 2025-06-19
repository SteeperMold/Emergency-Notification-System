package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"go.uber.org/zap"
)

func setupTemplateServer(tx domain.DBConn, userID any) *mux.Router {
	logger := zap.NewNop()
	timeout := 2 * time.Second

	r := mux.NewRouter()
	if userID != nil {
		r.Use(mockUserMiddleware(userID))
	}
	route.NewTemplateRoute(r, tx, logger, timeout)
	return r
}

type templateTC struct {
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

func TestTemplateHandler(t *testing.T) {
	cases := []templateTC{
		{
			name:           "Get_NoUserID",
			method:         http.MethodGet,
			path:           "/template",
			useTx:          false,
			expectedStatus: http.StatusInternalServerError,
			expectContains: "internal server error",
		},
		{
			name:   "Get_EmptyList",
			method: http.MethodGet,
			path:   "/template",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				now := time.Now().UTC().Truncate(time.Second)
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash,created_at) VALUES($1,$2,$3) RETURNING id`,
					"a@a.com", "h", now,
				).Scan(&uid)
				require.NoError(t, err)
				return uid, ""
			},
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},
		{
			name:   "Post_InvalidBody",
			method: http.MethodPost,
			path:   "/template",
			body:   map[string]string{"body": ""},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"b@b.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				return uid, ""
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectContains: "invalid template",
		},
		{
			name:   "Post_Success",
			method: http.MethodPost,
			path:   "/template",
			body:   map[string]string{"body": "Hello", "name": "name"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"c@c.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				return uid, ""
			},
			expectedStatus: http.StatusCreated,
			expectJSON:     true,
		},
		{
			name:           "GetById_BadID",
			method:         http.MethodGet,
			path:           "/template/xyz",
			useTx:          true,
			expectedStatus: http.StatusInternalServerError,
			expectContains: "internal server error",
		},
		{
			name:   "GetById_NotFound",
			method: http.MethodGet,
			path:   "/template/9999",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"d@d.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				return uid, ""
			},
			expectedStatus: http.StatusNotFound,
			expectContains: "template not exists",
		},
		{
			name:   "GetById_Success",
			method: http.MethodGet,
			path:   "/template/",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, tid int
				now := time.Now().UTC().Truncate(time.Second)
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"e@e.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				err = tx.QueryRow(ctx,
					`INSERT INTO message_templates(user_id,name,body,created_at,updated_at) VALUES($1,$2,$3,$4,$4) RETURNING id`,
					uid, "name", "T", now,
				).Scan(&tid)
				require.NoError(t, err)
				return uid, strconv.Itoa(tid)
			},
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},
		{
			name:   "Put_BadID",
			method: http.MethodPut,
			path:   "/template/abc",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				return 1, ""
			},
			expectedStatus: http.StatusBadRequest,
			expectContains: "invalid id",
		},
		{
			name:   "Put_InvalidBody",
			method: http.MethodPut,
			path:   "/template/",
			body:   map[string]string{"body": ""}, // too short
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, tid int
				// create user + template
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"x@x.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				err = tx.QueryRow(ctx,
					`INSERT INTO message_templates(user_id,name,body,created_at,updated_at)
					 VALUES($1,$2,$3,now(),now()) RETURNING id`,
					uid, "name", "orig",
				).Scan(&tid)
				require.NoError(t, err)
				return uid, strconv.Itoa(tid)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectContains: "invalid template",
		},
		{
			name:   "Put_NotFound",
			method: http.MethodPut,
			path:   "/template/",
			body:   map[string]string{"body": "NewBody", "name": "newName"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				// only user, no template
				var uid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"y@y.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				return uid, "9999"
			},
			expectedStatus: http.StatusNotFound,
			expectContains: "template not exists",
		},
		{
			name:   "Put_Success",
			method: http.MethodPut,
			path:   "/template/",
			body:   map[string]string{"body": "Updated", "name": "name"},
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, tid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"z@z.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				err = tx.QueryRow(ctx,
					`INSERT INTO message_templates(user_id,name,body,created_at,updated_at)
					 VALUES($1,$2,$3,now(),now()) RETURNING id`,
					uid, "name", "Original",
				).Scan(&tid)
				require.NoError(t, err)
				return uid, strconv.Itoa(tid)
			},
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},

		// DELETE -------------------------------------------------------------

		{
			name:   "Delete_BadID",
			method: http.MethodDelete,
			path:   "/template/abc",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				return 1, ""
			},
			expectedStatus: http.StatusBadRequest,
			expectContains: "invalid id",
		},
		{
			name:   "Delete_NotFound",
			method: http.MethodDelete,
			path:   "/template/",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"n@n.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				return uid, "9999"
			},
			expectedStatus: http.StatusNotFound,
			expectContains: "template not exists",
		},
		{
			name:   "Delete_Success",
			method: http.MethodDelete,
			path:   "/template/",
			useTx:  true,
			setup: func(ctx context.Context, tx domain.DBConn) (any, string) {
				var uid, tid int
				err := tx.QueryRow(ctx,
					`INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id`,
					"delt@d.com", "h",
				).Scan(&uid)
				require.NoError(t, err)
				err = tx.QueryRow(ctx,
					`INSERT INTO message_templates(user_id,name,body,created_at,updated_at)
					 VALUES($1,$2,$3,now(),now()) RETURNING id`,
					uid, "name", "ToDelete",
				).Scan(&tid)
				require.NoError(t, err)
				return uid, strconv.Itoa(tid)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func(ctx context.Context, tx domain.DBConn) {
				var userID any
				var tmplID string
				if tc.setup != nil {
					userID, tmplID = tc.setup(ctx, tx)
				}
				runTemplateTest(t, tx, tc, userID, tmplID)
			}

			if tc.useTx {
				testutils.WithRollback(t, run)
			} else {
				run(context.Background(), nil)
			}
		})
	}
}

func runTemplateTest(
	t *testing.T,
	tx domain.DBConn,
	tc templateTC,
	userID any,
	tmplID string,
) {
	router := setupTemplateServer(tx, userID)
	srv := httptest.NewServer(router)
	defer srv.Close()

	var bodyReader io.Reader
	if tc.body != nil {
		bs, _ := json.Marshal(tc.body)
		bodyReader = bytes.NewReader(bs)
	}

	url := srv.URL + tc.path
	if tmplID != "" {
		url += tmplID
	}

	req, err := http.NewRequest(tc.method, url, bodyReader)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	require.Equal(t, tc.expectedStatus, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	if tc.expectJSON {
		var v interface{}
		require.NoError(t, json.Unmarshal(b, &v))
	} else if tc.expectContains != "" {
		assert.Contains(t, string(b), tc.expectContains)
	}
}
