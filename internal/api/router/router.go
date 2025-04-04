package router

import (
	"net/http"
	"propagatorGo/internal/api/middleware"
	"propagatorGo/internal/api/response"
	"propagatorGo/internal/config"
	"propagatorGo/internal/repository"

	"github.com/gorilla/mux"
)

// Setup configures the main application router with all routes
func Setup(cfg *config.Config, articleRepo *repository.ArticleRepository) *mux.Router {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(middleware.CORS)

	// Create API subrouter with version prefix
	api := r.PathPrefix(cfg.App.APIPrefix).Subrouter()

	// Register route groups
	RegisterNewsRoutes(api, articleRepo)

	// Health check endpoint
	api.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

	// Not found handler
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	return r
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, map[string]string{
		"status":    "ok",
		"timestamp": http.TimeFormat,
		"version":   "1.0.0", // This should come from your app config
	}, http.StatusOK)
}

// notFoundHandler provides a custom 404 response
func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	response.NotFound(w, "The requested resource could not be found")
}
