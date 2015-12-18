package ligno

import (
	"runtime"
	"sync/atomic"
	"time"
)

// Logger is central datatype in ligno which represents logger itself.
// Logger is first level of processing events. It creates them and
// queues for async processing. It holds slice of Handlers that process
// messages and context (set of key-value pairs that will be include
// in every log record).
type Logger struct {
	// Context in which logger is operating. Basically, this is set of
	// key-value pairs that will be added to every record logged with this
	// logger. They have lowest priority.
	Context Record
	// Handlers is slice of handlers for processing messages.
	Handlers []Handler
	// messages is channel for queueing and buffering log records.
	messages  chan Record
	// toProcess is number of messages left to process in this logger.
	toProcess int32
}

// New creates new instance of logger and starts it so it is ready for processing.
func New(context Record, handlers []Handler) *Logger {
	l := &Logger{
		Context:  context,
		Handlers: handlers,
		messages: make(chan Record, 2048),
	}
	go l.run()
	return l
}

// run is log record processor which takes records from chan and invokes all handlers.
func (l *Logger) run() {
	var handlers []Handler
	if len(l.Handlers) == 0 {
		handlers = []Handler{stdoutHandler}
	} else {
		handlers = l.Handlers
	}
	for record := range l.messages {
		for _, h := range handlers {
			h.Handle(record)
		}
		atomic.AddInt32(&l.toProcess, -1)
	}
}

// wait blocks until all messages on messages channel are processed.
// Provided done channel will be closed when messages are processed to notify
// interested parties that they can unblock.
func (l *Logger) wait(done chan struct{}) {
	runtime.Gosched()
	for atomic.LoadInt32(&l.toProcess) > 0 {
		time.Sleep(1 * time.Millisecond)
	}
	close(done)
}

// Wait block until all messages sent to logger are processed.
// If timeout is needed, see WaitTimeout.
func (l *Logger) Wait() {
	done := make(chan struct{})
	go l.wait(done)
	<-done
}

// WaitTimeout blocks until all messages send to logger are processed or max
// specified amount of time.
// Boolean return value indicates if function returned because all messages
// were processed (true) or because timeout has expired (false).
func (l *Logger) WaitTimeout(t time.Duration) (ok bool) {
	done := make(chan struct{})
	timeout := time.After(t)
	go l.wait(done)
	select {
	case <-done:
		return true
	case <-timeout:
		return false
	}
}

// log creates record suitable for processing and sends it to messages chan.
func (l *Logger) log(data Record) {
	for k, v := range l.Context {
		data[k] = v
	}
	if _, ok := data[TIME]; !ok {
		data[TIME] = time.Now().UTC()
	}
	atomic.AddInt32(&l.toProcess, 1)
	l.messages <- data
}

// Log creates record and queues it for processing.
// Required parameters are level for record and event that occurred. Any
// additional parameters will be transformed to key-value pairs for record
// in order in which they were provided. There should be even number of them,
// but in case that there is on number of parameters, empty string is
// appended. Example:
//   l.Log(INFO, "User logged in", "user_id", user_id, "platform", PLATFORM_NAME)
// will be translated into log record with following keys:
//  {LEVEL: INFO", EVENT: "User logged in", "user_id": user_id, "platform": PLATFORM_NAME}
func (l *Logger) Log(level Level, event string, pairs ...string) {
	var d = make(Record)
	d[EVENT] = event
	d[LEVEL] = level
	// make sure that number of items in data is even
	if len(pairs)%2 != 0 {
		pairs = append(pairs, "")
	}
	for i := 0; i < len(pairs); i += 2 {
		d[pairs[i]] = pairs[i+1]
	}
	l.log(d)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Debug(event string, pairs ...string) {
	l.Log(DEBUG, event, pairs...)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Info(event string, pairs ...string) {
	l.Log(INFO, event, pairs...)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Warning(event string, pairs ...string) {
	l.Log(WARNING, event, pairs...)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Error(event string, pairs ...string) {
	l.Log(ERROR, event, pairs...)
}
