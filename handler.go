package ligno

import (
	"io"
	"log/syslog"
	"os"
	"sync"
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

// StreamHandler writes records to provided io.Writer
func StreamHandler(out io.Writer, formatter Formatter) Handler {
	return HandlerFunc(func(record Record) error {
		_, err := out.Write(formatter.Format(record))
		return err
	})
}

// Predicate is function that returns true if record should be logged, false otherwise.
type Predicate func(Record) bool

// FilterHandler checks records if by using predicate to check if they should
// be processed and only if they do, record is passed to provided handler.
func FilterHandler(predicate Predicate, handler Handler) Handler {
	return HandlerFunc(func(record Record) error {
		if predicate(record) {
			return handler.Handle(record)
		}
		return nil
	})
}

// FilterLevelHandler is FilterHandler with default predicate function that filters
// all records below provided level.
func FilterLevelHandler(level Level, handler Handler) Handler {
	levelPredicate := Predicate(func(record Record) bool {
		return record.Level >= level
	})
	return FilterHandler(levelPredicate, handler)
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

// FileHandler writes log records to file with provided name.
func FileHandler(fileName string, formatter Formatter) Handler {
	return &fileHandler{
		fileName:  fileName,
		formatter: formatter,
	}
}

// fileHandler writes log messages to file with provided name.
type fileHandler struct {
	fileName  string
	formatter Formatter
	f         *os.File
}

// Handle writes record to file.
func (fh *fileHandler) Handle(record Record) error {
	if fh.f == nil {
		f, err := os.OpenFile(fh.fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		fh.f = f
	}

	_, err := fh.f.Write(fh.formatter.Format(record))
	return err
}

// Close closes file were records are being written.
func (fh *fileHandler) Close() {
	fh.f.Close()
}

// NullHandler returns handler that discards all records.
func NullHandler() Handler {
	return HandlerFunc(func(record Record) error {
		return nil
	})
}

// InspectHandler is handler that is able to restore logged message and
// return them for inspection.
type InspectHandler interface {
	Handler
	Messages() []string
}

// MemoryHandler stores all records in memory, to be fetched and inspected later.
type memoryHandler struct {
	buffer    [][]byte
	formatter Formatter
	mu        sync.Mutex
}

// Handle stores formatted record in memory.
func (mh *memoryHandler) Handle(record Record) error {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	mh.buffer = append(mh.buffer, mh.formatter.Format(record))
	return nil
}

// Content returns content sent to logging to memory handler.
func (mh *memoryHandler) Messages() []string {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	messages := make([]string, 0, len(mh.buffer))
	for _, msg := range mh.buffer {
		messages = append(messages, string(msg))
	}
	return messages
}

// MemoryHandler returns handler instance that saves all message to memory.
func MemoryHandler(formatter Formatter) InspectHandler {
	return &memoryHandler{
		buffer:    make([][]byte, 0),
		formatter: formatter,
	}
}

// syslogHandler sends all messages to local syslog server.
type syslogHandler struct {
	Formatter Formatter
	Tag       string
	Priority  syslog.Priority
	writer    *syslog.Writer
}

// SyslogHandler creates new syslog handler with provided config variables.
func SyslogHandler(formatter Formatter, tag string, priority syslog.Priority) Handler {
	writer, err := syslog.New(syslog.LOG_DEBUG, tag)
	if err != nil {
		panic(err)
	}
	return &syslogHandler{
		Formatter: formatter,
		Tag:       tag,
		Priority:  priority,
		writer:    writer,
	}
}

// Handle passes all messages to syslog server. Message priorities are
// translated to syslog compatible priorities.
func (sh *syslogHandler) Handle(record Record) error {

	msg := string(sh.Formatter.Format(record))
	switch record.Level {
	case NOTSET:
		return sh.writer.Info(msg)
	case DEBUG:
		return sh.writer.Debug(msg)
	case INFO:
		return sh.writer.Info(msg)
	case WARNING:
		return sh.writer.Warning(msg)
	case ERROR:
		return sh.writer.Err(msg)
	case CRITICAL:
		return sh.writer.Crit(msg)
	default:
		return sh.writer.Info(msg)
	}
	return nil
}

// Close closes connection with syslog server.
func (sh *syslogHandler) Close() {
	sh.writer.Close()
}
