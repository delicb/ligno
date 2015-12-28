package ligno

import "time"

// Ctx is additional context for log record.
type Ctx map[string]interface{}

// merge creates new context, merges this one with provided one and returns it.
func (ctx Ctx) merge(other Ctx) (merged Ctx) {
	newCtx := make(Ctx)
	for k, v := range ctx {
		newCtx[k] = v
	}
	for k, v := range other {
		newCtx[k] = v
	}
	return newCtx
}

// Record holds information about one log message.
type Record struct {
	Time    time.Time `json:"time"`
	Level   Level     `json:"level"`
	Message string    `json:"message"`
	Context Ctx       `json:"context"`
	Logger  *Logger   `json:"-"`
}
