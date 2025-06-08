package tokenutils_test

import (
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/internal/tokenutils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const secret = "supersecret"

func newTestUser() *models.User {
	return &models.User{
		ID:    123,
		Email: "alice@example.com",
	}
}

type jwtTestCase struct {
	desc     string
	tokenGen func() (string, error)
	extract  func(string) (int, error)
	expectID int
	expectOk bool
}

func TestCreateToken_And_ExtractID(t *testing.T) {
	user := newTestUser()
	tests := []jwtTestCase{
		{
			desc: "valid access token",
			tokenGen: func() (string, error) {
				return tokenutils.CreateAccessToken(user, secret, time.Minute)
			},
			extract: func(tok string) (int, error) {
				return tokenutils.ExtractIDFromToken(tok, secret)
			},
			expectID: user.ID,
			expectOk: true,
		},
		{
			desc: "valid refresh token",
			tokenGen: func() (string, error) {
				return tokenutils.CreateRefreshToken(user, secret, time.Minute)
			},
			extract: func(tok string) (int, error) {
				return tokenutils.ExtractIDFromToken(tok, secret)
			},
			expectID: user.ID,
			expectOk: true,
		},
		{
			desc: "invalid token string",
			tokenGen: func() (string, error) {
				return "notatoken", nil
			},
			extract: func(tok string) (int, error) {
				return tokenutils.ExtractIDFromToken(tok, secret)
			},
			expectID: 0,
			expectOk: false,
		},
		{
			desc: "missing id claim",
			tokenGen: func() (string, error) {
				claims := jwt.MapClaims{
					"email": user.Email,
					"exp":   time.Now().Add(time.Minute).Unix(),
					"iat":   time.Now().Unix(),
				}
				t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				return t.SignedString([]byte(secret))
			},
			extract: func(tok string) (int, error) {
				return tokenutils.ExtractIDFromToken(tok, secret)
			},
			expectID: 0,
			expectOk: false,
		},
		{
			desc: "id wrong type",
			tokenGen: func() (string, error) {
				claims := jwt.MapClaims{
					"id":  "not-a-number",
					"exp": time.Now().Add(time.Minute).Unix(),
				}
				t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				return t.SignedString([]byte(secret))
			},
			extract: func(tok string) (int, error) {
				return tokenutils.ExtractIDFromToken(tok, secret)
			},
			expectID: 0,
			expectOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tok, err := tc.tokenGen()
			require.NoError(t, err)

			id, err := tc.extract(tok)
			if tc.expectOk {
				require.NoError(t, err)
				assert.Equal(t, tc.expectID, id)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectID, id)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	user := newTestUser()
	t.Run("valid", func(t *testing.T) {
		tok, err := tokenutils.CreateAccessToken(user, secret, time.Minute)
		require.NoError(t, err)
		auth, err := tokenutils.IsAuthorized(tok, secret)
		assert.NoError(t, err)
		assert.True(t, auth)
	})
	t.Run("invalid signature", func(t *testing.T) {
		tok, err := tokenutils.CreateAccessToken(user, "wrongsecret", time.Minute)
		require.NoError(t, err)
		auth, err := tokenutils.IsAuthorized(tok, secret)
		assert.Error(t, err)
		assert.False(t, auth)
	})
}
