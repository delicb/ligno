package ligno

import "time"

// Ctx is additional context for log record.
type Ctx map[string]interface{}

// Record holds information about one log message.
type Record struct {
	Time    time.Time `json:"time"`
	Level   Level     `json:"level"`
	Message string    `json:"message"`
	Context Ctx       `json:"context"`
	Logger  *Logger   `json:"-"`
}
