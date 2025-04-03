package handlers

import (
	"net/http"
	"propagatorGo/internal/api/response"
	"strconv"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct{}

// GetLimitParam extracts and validates a limit parameter from the request
func (h *BaseHandler) GetLimitParam(r *http.Request, defaultLimit, maxLimit int) int {
	limitStr := r.URL.Query().Get("limit")
	limit := defaultLimit

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			if maxLimit > 0 && parsedLimit > maxLimit {
				limit = maxLimit
			} else {
				limit = parsedLimit
			}
		}
	}

	return limit
}

// GetPageParam extracts and validates a page parameter from the request
func (h *BaseHandler) GetPageParam(r *http.Request, defaultPage int) int {
	pageStr := r.URL.Query().Get("page")
	page := defaultPage

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	return page
}

// Paginate creates a paginated response structure
func (h *BaseHandler) Paginate(data interface{}, total, perPage, page int) response.PaginatedResponse {
	lastPage := (total + perPage - 1) / perPage // Ceiling division

	return response.PaginatedResponse{
		Data: data,
		Pagination: response.Pagination{
			Total:       total,
			PerPage:     perPage,
			CurrentPage: page,
			LastPage:    lastPage,
		},
	}
}
