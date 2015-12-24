package ligno

import (
	"fmt"
	"io"
	"os"
)

// Handler processes log records and writes them to appropriate destination.
type Handler interface {
	// Handle processes provided log record.
	Handle(Record) error
}

// HandlerCloser is interface that allows handlers to be closed.
// If handler implements this interface, when logger is stopped, Close will
// be called.
type HandlerCloser interface {
	Close()
}

// HandlerFunc is function that implements Handler interface.
type HandlerFunc func(Record) error

// Handle just calls HandlerFunc.
func (hf HandlerFunc) Handle(record Record) error {
	return hf(record)
}


func StreamHandler(out io.Writer, formatter Formatter) Handler {
	return HandlerFunc(func(record Record) error {
		_, err := out.Write([]byte(formatter.Format(record)))
		return err
	})
}

// combiningHandler combines multiple other handlers
type combiningHandler struct {
	Handlers []Handler
}

// Handle processes record by passing it to all internal handler of this handler.
func (ch *combiningHandler) Handle(record Record) error {
	var err error
	for _, h := range ch.Handlers {
		err = h.Handle(record)
	}
	return err
}

// Close closes all internal handlers if they implement HandlerCloser interface.
func (ch *combiningHandler) Close() {
	for _, h := range ch.Handlers {
		if handlerCloser, ok := h.(HandlerCloser); ok {
			handlerCloser.Close()
		}
	}
}

// CombiningHandler creates and returns handler that passes records to all
// provided handlers.
func CombiningHandler(handlers ...Handler) Handler {
	return &combiningHandler{
		Handlers: handlers,
	}
}

// Filter determines if log record should be processed further.
type Filter interface {
	// ShouldLog returns bool flag that indicates whether message should be
	// processed or discarded.
	ShouldLog(Record) bool
}

// LevelFilter determines if message should be processed or discarded based
// on level in log record.
type LevelFilter struct {
	// Minimal log level that log record has to have in order to pass this filter.
	Level Level
}

// ShouldLog returns true if record level is higher or equal to on set in this LevelFilter.
func (f LevelFilter) ShouldLog(r Record) bool {
	should := r.Level >= f.Level
	return should
}

// shouldLog determines if provided record should be processed based on provided filters.
// If no filters are provided, true will be returned. Otherwise, if at least one
// filter determined that record should be logged, true is returned, false otherwise.
func shouldLog(record Record, filters []Filter) bool {
	// edge case - if there are not filters, message should be logged
	if len(filters) == 0 {
		return true
	}
	res := false
	for _, f := range filters {
		res = res || f.ShouldLog(record)
	}
	return res
}

var _out io.Writer = os.Stdout

// StdoutHandler is handler that prints log records to standard output.
type StdoutHandler struct {
	// Formatter that will be used to format record strings before printing it.
	Formatter Formatter
	// Filters to determine if log record should be logged.
	Filters []Filter
}

// Handle processes provided log records by printing it to standard output.
func (sh *StdoutHandler) Handle(record Record) error {
	if !shouldLog(record, sh.Filters) {
		return nil
	}
	var formatter Formatter
	if sh.Formatter == nil {
		formatter = defaultFormatter
	} else {
		formatter = sh.Formatter
	}
	_, err := fmt.Fprintln(_out, formatter.Format(record))
	return err
}

// stdoutHandler is default instance of StdoutHandler
var stdoutHandler = &StdoutHandler{}
