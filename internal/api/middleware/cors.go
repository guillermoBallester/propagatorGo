package middleware

import (
	"net/http"
)

// CORS middleware adds Cross-Origin Resource Sharing headers to responses
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// CORSWithConfig is a more configurable version of the CORS middleware
func CORSWithConfig(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	// Default values if not provided
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type", "Authorization"}
	}

	// Join arrays into comma-separated strings
	origins := joinStrings(allowedOrigins, ", ")
	methods := joinStrings(allowedMethods, ", ")
	headers := joinStrings(allowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// joinStrings joins an array of strings with a separator
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}

	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}
