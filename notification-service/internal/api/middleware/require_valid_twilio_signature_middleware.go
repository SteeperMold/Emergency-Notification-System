package middleware

import (
	"github.com/twilio/twilio-go/client"
	"net/http"
)

func RequireValidTwilioSignatureMiddleware(baseUrl string, validator *client.RequestValidator) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := baseUrl + r.URL.Path
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
