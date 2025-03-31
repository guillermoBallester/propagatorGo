package task

import (
	"encoding/json"
	"fmt"
	"propagatorGo/internal/constants"
	scraper "propagatorGo/internal/scrapper"
	"time"
)

// Params contains all necessary parameters for a task
type Params struct {
	// Generic fields
	Symbol   string                 `json:"symbol,omitempty"`
	URL      string                 `json:"url,omitempty"`
	Source   string                 `json:"source,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Task represents a unit of work in the system
type Task struct {
	ID        string    `json:"id,omitempty"`
	Type      string    `json:"type"`
	Params    Params    `json:"params"`
	Priority  int       `json:"priority,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTask creates a task
func NewTask(symbol string, taskType string, source string) *Task {
	return &Task{
		Type: taskType,
		Params: Params{
			Symbol: symbol,
			Source: source,
		},
		CreatedAt: time.Now(),
	}
}

// GetArticle extracts the article from a consume task
func (t *Task) GetArticle() (*scraper.ArticleData, error) {
	if t.Type != constants.TaskTypeConsume {
		return nil, fmt.Errorf("task is not a consume task")
	}

	articleData, ok := t.Params.Metadata["article"]
	if !ok {
		return nil, fmt.Errorf("article data not found in task")
	}

	// Convert the interface{} back to a map
	articleMap, ok := articleData.(map[string]interface{})
	if !ok {
		// If it's not already a map, it might be stored as JSON
		articleJSON, err := json.Marshal(articleData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal article data: %w", err)
		}

		var article scraper.ArticleData
		if err := json.Unmarshal(articleJSON, &article); err != nil {
			return nil, fmt.Errorf("failed to unmarshal article data: %w", err)
		}

		return &article, nil
	}

	// Convert the map to an Article struct
	articleJSON, err := json.Marshal(articleMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal article map: %w", err)
	}

	var article scraper.ArticleData
	if err := json.Unmarshal(articleJSON, &article); err != nil {
		return nil, fmt.Errorf("failed to unmarshal article map: %w", err)
	}

	return &article, nil
}
