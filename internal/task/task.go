package task

import "time"

// Params contains all necessary parameters for a task
type Params struct {
	// Generic fields
	Symbol   string            `json:"symbol,omitempty"`
	URL      string            `json:"url,omitempty"`
	Source   string            `json:"source,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
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
