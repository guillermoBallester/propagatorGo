package response

import (
	"net/http"
)

// ErrorResponse provides a standardized error response structure
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Error sends a standardized error response
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, ErrorResponse{
		Status:  statusCode,
		Message: message,
	}, statusCode)
}

// ErrorWithCode sends a standardized error response with an error code
func ErrorWithCode(w http.ResponseWriter, statusCode int, message string, code string) {
	JSON(w, ErrorResponse{
		Status:  statusCode,
		Message: message,
		Code:    code,
	}, statusCode)
}

// ValidationError represents an error in request validation
type ValidationError struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors"`
}

// ValidationErrors sends a response for validation errors
func ValidationErrors(w http.ResponseWriter, errors map[string]interface{}) {
	JSON(w, ValidationError{
		Status:  http.StatusUnprocessableEntity,
		Message: "The request contains invalid parameters",
		Errors:  errors,
	}, http.StatusUnprocessableEntity)
}

// NotFound sends a standardized not found response
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, http.StatusNotFound, message)
}

// BadRequest sends a standardized bad request response
func BadRequest(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Bad request"
	}
	Error(w, http.StatusBadRequest, message)
}

// InternalServerError sends a standardized internal server error response
func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(w, http.StatusInternalServerError, message)
}

// Unauthorized sends a standardized unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden sends a standardized forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Error(w, http.StatusForbidden, message)
}
