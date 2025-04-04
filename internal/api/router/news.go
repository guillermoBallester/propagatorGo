package router

import (
	"net/http"
	"propagatorGo/internal/api/handlers"
	"propagatorGo/internal/repository"

	"github.com/gorilla/mux"
)

// RegisterNewsRoutes sets up all news-related routes
func RegisterNewsRoutes(r *mux.Router, articleRepo *repository.ArticleRepository) {
	newsHandler := handlers.NewNewsHandler(articleRepo)

	// GET /stocks/{symbol}/news - News for a specific stock
	r.HandleFunc("/stocks/{symbol}/news", newsHandler.GetBySymbol).Methods(http.MethodGet)

	// GET /sources/{site}/news - News from a specific source
	r.HandleFunc("/sources/{site}/news", newsHandler.GetBySite).Methods(http.MethodGet)
}
