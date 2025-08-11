package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	apiBaseURL   = "http://apiservice:8080"
	signupPath   = "/signup"
	templatePath = "/template"
	contactPath  = "/contacts"
	sendPath     = "/send-notification"
)

const (
	ntfsCountQuery = `
		SELECT count(*)
		FROM notifications
	`
	sentNtfsCountQuery = `
		SELECT count(*)
		FROM notifications
		WHERE status = 'sent'
	`
)

func TestBasicNotificationFlow(t *testing.T) {
	row := testPool.QueryRow(context.Background(), ntfsCountQuery)
	var initialCount int
	err := row.Scan(&initialCount)
	require.NoError(t, err)

	client := &http.Client{}

	signupBytes, err := json.Marshal(map[string]any{
		"email":    "test@e2e.com",
		"password": "123456789admin",
	})
	require.NoError(t, err)

	signupResp, err := client.Post(apiBaseURL+signupPath, "application/json", bytes.NewReader(signupBytes))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, signupResp.StatusCode)
	defer func() {
		err := signupResp.Body.Close()
		require.NoError(t, err)
	}()

	var signupResult map[string]any
	err = json.NewDecoder(signupResp.Body).Decode(&signupResult)
	require.NoError(t, err)
	accessToken, ok := signupResult["accessToken"].(string)
	require.True(t, ok)
	require.NotEmpty(t, accessToken)

	templateBytes, err := json.Marshal(map[string]any{
		"name": "Basic test template name",
		"body": "Basic test template body",
	})
	require.NoError(t, err)

	templateReq, err := http.NewRequest(http.MethodPost, apiBaseURL+templatePath, bytes.NewReader(templateBytes))
	require.NoError(t, err)
	templateReq.Header.Set("Authorization", "Bearer "+accessToken)
	templateResp, err := client.Do(templateReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, templateResp.StatusCode)
	defer func() {
		err := templateResp.Body.Close()
		require.NoError(t, err)
	}()

	var templateResult map[string]any
	err = json.NewDecoder(templateResp.Body).Decode(&templateResult)
	require.NoError(t, err)
	templateID, ok := templateResult["id"].(float64)
	require.True(t, ok)
	require.NotEmpty(t, templateID)
	templateIDStr := strconv.Itoa(int(templateID))

	contactBytes, err := json.Marshal(map[string]any{
		"name":  "Test contact name",
		"phone": "+79123456789",
	})
	require.NoError(t, err)

	contactsReq, err := http.NewRequest(http.MethodPost, apiBaseURL+contactPath, bytes.NewReader(contactBytes))
	require.NoError(t, err)
	contactsReq.Header.Set("Authorization", "Bearer "+accessToken)
	contactsResp, err := client.Do(contactsReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, contactsResp.StatusCode)
	defer func() {
		err := contactsResp.Body.Close()
		require.NoError(t, err)
	}()

	sendReq, err := http.NewRequest(http.MethodPost, apiBaseURL+sendPath+"/"+templateIDStr, nil)
	require.NoError(t, err)
	sendReq.Header.Set("Authorization", "Bearer "+accessToken)
	sendResp, err := client.Do(sendReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, sendResp.StatusCode)
	defer func() {
		err := sendResp.Body.Close()
		require.NoError(t, err)
	}()

	deadline := time.Now().Add(30 * time.Second)
	ntfCreated := false

	for time.Now().Before(deadline) {
		row = testPool.QueryRow(context.Background(), sentNtfsCountQuery)
		var count int
		err = row.Scan(&count)
		require.NoError(t, err)

		if count-initialCount == 1 {
			ntfCreated = true
			break
		}
	}

	require.True(t, ntfCreated)
}
