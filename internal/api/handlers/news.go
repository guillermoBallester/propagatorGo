package handlers

import (
	"net/http"
	"propagatorGo/internal/api/response"
	"propagatorGo/internal/database"
	"propagatorGo/internal/repository"
	"time"

	"github.com/gorilla/mux"
)

// NewsHandler handles news-related API requests
type NewsHandler struct {
	BaseHandler
	articleRepo *repository.ArticleRepository
}

// ArticleResponse represents the article data sent to the client
type ArticleResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Text      string    `json:"text,omitempty"`
	SiteName  string    `json:"site_name"`
	Symbol    string    `json:"symbol"`
	ScrapedAt time.Time `json:"scraped_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(repo *repository.ArticleRepository) *NewsHandler {
	return &NewsHandler{
		articleRepo: repo,
	}
}

// GetBySymbol handles requests for articles by stock symbol
func (h *NewsHandler) GetBySymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	// Get pagination parameters
	limit := h.GetLimitParam(r, 10, 50)
	page := h.GetPageParam(r, 1)

	articles, total, err := h.articleRepo.GetArticlesBySymbol(r.Context(), symbol)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error retrieving articles")
		return
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit
	if start >= len(articles) {
		// Return empty array if page is beyond available data
		response.JSON(w, h.Paginate([]ArticleResponse{}, total, limit, page), http.StatusOK)
		return
	}
	if end > len(articles) {
		end = len(articles)
	}
	paginatedArticles := articles[start:end]

	// Map to response objects
	responses := mapArticlesToResponse(paginatedArticles)
	response.JSON(w, h.Paginate(responses, total, limit, page), http.StatusOK)
}

// GetBySite handles requests for articles by site name
func (h *NewsHandler) GetBySite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteName := vars["site"]

	// Get pagination parameters
	limit := h.GetLimitParam(r, 10, 50)
	page := h.GetPageParam(r, 1)

	articles, total, err := h.articleRepo.GetArticlesBySite(r.Context(), siteName)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error retrieving articles")
		return
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit
	if start >= len(articles) {
		// Return empty array if page is beyond available data
		response.JSON(w, h.Paginate([]ArticleResponse{}, total, limit, page), http.StatusOK)
		return
	}
	if end > len(articles) {
		end = len(articles)
	}
	paginatedArticles := articles[start:end]

	// Map to response objects
	responses := mapArticlesToResponse(paginatedArticles)
	response.JSON(w, h.Paginate(responses, total, limit, page), http.StatusOK)
}

// Helper function to map a database article to API response
func mapArticleToResponse(article database.Article) ArticleResponse {
	return ArticleResponse{
		ID:        article.ID,
		Title:     article.Title,
		URL:       article.URL,
		Text:      article.Text,
		SiteName:  article.SiteName,
		Symbol:    article.Symbol,
		ScrapedAt: article.ScrapedAt,
		CreatedAt: article.CreatedAt,
	}
}

// Helper function to map multiple articles to responses
func mapArticlesToResponse(articles []database.Article) []ArticleResponse {
	responses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		responses[i] = mapArticleToResponse(article)
	}
	return responses
}
