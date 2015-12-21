package ligno

import (
	"fmt"
	"os"
	"io"
)

// Handler processes log records and writes them to appropriate destination.
type Handler interface {
	// Handle processes provided log record.
	Handle(Record)
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
	should := r.Level() >= f.Level
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
func (sh *StdoutHandler) Handle(record Record) {
	if !shouldLog(record, sh.Filters) {
		return
	}
	var formatter Formatter
	if sh.Formatter == nil {
		formatter = defaultFormatter
	} else {
		formatter = sh.Formatter
	}
	fmt.Fprintln(_out, formatter.Format(record))
}

// stdoutHandler is default instance of StdoutHandler
var stdoutHandler = &StdoutHandler{}
