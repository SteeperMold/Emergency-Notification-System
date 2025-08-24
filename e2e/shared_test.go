package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	apiBaseURL       = "http://apiservice:8080"
	signupPath       = "/signup"
	templatePath     = "/templates"
	contactPath      = "/contacts"
	loadContactsPath = "/load-contacts"
	sendPath         = "/send-notification"
)

const sentNtfsCountQuery = `
	SELECT count(*)
	FROM notifications
	WHERE status = 'sent'
`

func FetchSentNotificationsCount(t *testing.T) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := testPool.QueryRow(ctx, sentNtfsCountQuery)

	var count int
	err := row.Scan(&count)
	require.NoError(t, err)

	return count
}

func SendSignupRequest(t *testing.T, email, password string) (accessToken string) {
	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	require.NoError(t, err)

	resp, err := http.Post(apiBaseURL+signupPath, "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	var res map[string]any
	err = json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(t, err)

	accessToken, ok := res["accessToken"].(string)
	require.True(t, ok)
	require.NotEmpty(t, accessToken)

	return accessToken
}

func PostTemplateRequest(t *testing.T, accessToken, name, body string) (id int) {
	reqBody, err := json.Marshal(map[string]string{
		"name": name,
		"body": body,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, apiBaseURL+templatePath, bytes.NewReader(reqBody))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	var res map[string]any
	err = json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(t, err)

	templateID, ok := res["id"].(float64)
	require.True(t, ok)
	require.NotEmpty(t, templateID)

	return int(templateID)
}

func PostContactRequest(t *testing.T, accessToken, name, phone string) {
	reqBody, err := json.Marshal(map[string]string{
		"name":  name,
		"phone": phone,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, apiBaseURL+contactPath, bytes.NewReader(reqBody))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
}

func GetContactsRequest(t *testing.T, accessToken string) (contactsCount int) {
	req, err := http.NewRequest(http.MethodGet, apiBaseURL+contactPath, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	var res map[string]any
	err = json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(t, err)

	total, ok := res["total"].(float64)
	require.True(t, ok)

	return int(total)
}

func SendLoadContactsFromFileRequest(t *testing.T, accessToken, filePath string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer func() {
		err = file.Close()
		require.NoError(t, err)
	}()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", apiBaseURL+loadContactsPath, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func SendNotificationRequest(t *testing.T, accessToken string, templateID int) {
	req, err := http.NewRequest(http.MethodPost, apiBaseURL+sendPath+"/"+strconv.Itoa(templateID), nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
}
