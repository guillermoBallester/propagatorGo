package task

import (
	"encoding/json"
	"fmt"
	"propagatorGo/internal/constants"
	scraper "propagatorGo/internal/scrapper"
	"time"
)

// Task represents a unit of work in the system
type Task struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"type"`
	Params    map[string]interface{} `json:"params"`
	Priority  int                    `json:"priority,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewTask creates a new task with basic parameters
func NewTask(taskType string) *Task {
	return &Task{
		Type:      taskType,
		Params:    make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// SetParam sets a parameter for the task
func (t *Task) SetParam(key string, value interface{}) {
	t.Params[key] = value
}

// GetParam retrieves a parameter from the task
func (t *Task) GetParam(key string) (interface{}, bool) {
	value, exists := t.Params[key]
	return value, exists
}

// GetParamString retrieves a string parameter from the task
func (t *Task) GetParamString(key string) (string, error) {
	value, exists := t.Params[key]
	if !exists {
		return "", fmt.Errorf("parameter %s not found", key)
	}

	str, ok := value.(string)
	if !ok {
		// Try to convert to string if it's not already
		return fmt.Sprintf("%v", value), nil
	}
	return str, nil
}

// GetArticle extracts the article from a consume task
func (t *Task) GetArticle() (*scraper.ArticleData, error) {
	if t.Type != constants.TaskTypeConsume {
		return nil, fmt.Errorf("task is not a consume task")
	}

	articleData, ok := t.Params["article"]
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
