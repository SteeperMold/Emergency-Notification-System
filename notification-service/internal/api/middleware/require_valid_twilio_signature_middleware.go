package middleware

import (
	"net/http"

	"github.com/twilio/twilio-go/client"
)

// RequireValidTwilioSignatureMiddleware returns a middleware that validates incoming Twilio
// webhook requests by verifying the X-Twilio-Signature header against the request URL and parameters.
// This ensures that only genuine requests from Twilio are processed.
func RequireValidTwilioSignatureMiddleware(baseURL string, validator *client.RequestValidator) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := baseURL + r.URL.Path
			signatureHeader := r.Header.Get("X-Twilio-Signature")
			params := make(map[string]string)

			err := r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			for key, value := range r.PostForm {
				params[key] = value[0]
			}

			if !validator.Validate(url, params, signatureHeader) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
