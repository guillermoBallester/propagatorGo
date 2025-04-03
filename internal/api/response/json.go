package response

import (
	"encoding/json"
	"net/http"
)

// PaginatedResponse provides a standard structure for paginated responses
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination holds pagination metadata
type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}

// JSON writes a JSON response with appropriate headers
func JSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For CORS
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// JSONWithHeaders writes a JSON response with custom headers
func JSONWithHeaders(w http.ResponseWriter, data interface{}, statusCode int, headers map[string]string) {
	// Set content type first
	w.Header().Set("Content-Type", "application/json")

	// Set additional headers
	for key, value := range headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
