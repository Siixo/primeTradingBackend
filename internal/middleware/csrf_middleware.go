package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

const CSRFCookieName = "csrf_token"
const CSRFHeaderName = "X-CSRF-Token"

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CSRFCookieName)
		var token string
		
		if err != nil || cookie.Value == "" {
			token = generateToken()
			http.SetCookie(w, &http.Cookie{
				Name:     CSRFCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: false, // Frontend JS needs to read this cookie to send in header
				Secure:   true,  // Mandatory for cross-origin (Vercel to Oracle VM)
				SameSite: http.SameSiteNoneMode,
			})
		} else {
			token = cookie.Value
		}

		// Safe methods do not require CSRF validation
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// State-changing methods require the token in the header
		headerToken := r.Header.Get(CSRFHeaderName)
		if headerToken == "" || headerToken != token {
			writeJSONError(w, "Invalid or missing CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
