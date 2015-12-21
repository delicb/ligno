package ligno

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type tracker struct {
	loggers       []*Logger
	loggerCreated chan *Logger
	mu            sync.RWMutex
}

func (lt *tracker) run() {
	for l := range lt.loggerCreated {
		lt.mu.Lock()
		lt.loggers = append(lt.loggers, l)
		lt.mu.Unlock()
	}
}

func (lt *tracker) wait(done chan struct{}) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	var wg sync.WaitGroup
	for _, l := range lt.loggers {
		wg.Add(1)
		go func(l *Logger) {
			l.Wait()
			wg.Done()
		}(l)
	}
	wg.Wait()
	close(done)
}

var loggerTracker = tracker{
	loggers:       make([]*Logger, 0, 8),
	loggerCreated: make(chan *Logger),
}

var defaultLog *Logger

func init() {
	go loggerTracker.run()
	// done here instead of in default to make sure that tracker is
	// running before creating any loggers
	defaultLog = New(nil, []Handler{stdoutHandler})
}

// WaitAll blocks until all loggers are finished with message processing.
func WaitAll() {
	done := make(chan struct{})
	loggerTracker.wait(done)
	<-done
}

// WaitAllTimeout blocks until all messages send to all loggers are processed or max
// specified amount of time.
// Boolean return value indicates if function returned because all messages
// were processed (true) or because timeout has expired (false).
func WaitAllTimeout(t time.Duration) bool {
	done := make(chan struct{})
	timeout := time.After(t)
	go loggerTracker.wait(done)
	select {
	case <-done:
		return true
	case <-timeout:
		return false
	}
}

// Logger is central data type in ligno which represents logger itself.
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
	// Parent is logger that this logger was created from.
	// It is used to create hierarchy of Context values.
	Parent *Logger
	// rawMessages is channel for queueing and buffering raw messages from
	// application which needs to be merged with context and submitted
	// to final processing
	rawRecords chan Record
	// records is channel for queueing and buffering log records.
	records chan Record
	// propagate is flag that determines if records will be propagated to
	// parent logger for processing too.
	propagate bool
	// notifyFinished is channel of channels. When someone wants to be notified
	// when logger processed all queued records, it sends channel that will be
	// closed after last queued record is processed to notifyFinished.
	notifyFinished chan chan struct{}
	// toProcess is number of messages left to process in this logger.
	toProcess int32
}

// New creates new instance of logger and starts it so it is ready for processing.
func New(context Record, handlers []Handler) *Logger {
	l := &Logger{
		Context:        context,
		Handlers:       handlers,
		records:        make(chan Record, 2048),
		rawRecords:     make(chan Record, 2048),
		notifyFinished: make(chan chan struct{}),
	}
	go l.handle()
	go l.processRecords()
	loggerTracker.loggerCreated <- l
	return l
}

// SubLogger creates and returns new logger whose parent is current logger.
func (l *Logger) SubLogger(context Record, handlers []Handler, propagate bool) *Logger {
	newLogger := New(context, handlers)
	newLogger.Parent = l
	newLogger.propagate = propagate
	return newLogger
}

// handle is log record processor which takes records from chan and invokes all handlers.
func (l *Logger) handle() {
	var handlers []Handler
	if len(l.Handlers) == 0 {
		handlers = []Handler{stdoutHandler}
	} else {
		handlers = l.Handlers
	}

	var notifyFinished chan struct{}
	for {
		select {
		case record := <-l.records:
			for _, h := range handlers {
				h.Handle(record)
			}
			atomic.AddInt32(&l.toProcess, -1)
			// if count dropped to 0, close notification channel
			if atomic.LoadInt32(&l.toProcess) == 0 && notifyFinished != nil {
				close(notifyFinished)
				// reset notification channel
				notifyFinished = nil
			}
		case notifyFinished = <-l.notifyFinished:
			// check count right away and notify that processing is done if possible
			if atomic.LoadInt32(&l.toProcess) == 0 {
				close(notifyFinished)
				// reset notification channel
				notifyFinished = nil
			}
		}
	}
}

// processRecords creates full records from provided user record and this and
// all parents contexts.
func (l *Logger) processRecords() {
	for data := range l.rawRecords {
		var current = l
		var loggerChain = make([]*Logger, 0, 5)

		// create list of all parents
		for current != nil {
			loggerChain = append(loggerChain, current)
			current = current.Parent
		}
		// merge context of all parents backwards, because children can override parents
		mergedData := make(Record)
		for i := len(loggerChain) - 1; i >= 0; i-- {
			for k, v := range loggerChain[i].Context {
				mergedData[k] = v
			}
		}
		// finally, add provided data to override all context keys
		for k, v := range data {
			mergedData[k] = v
		}
		l.records <- mergedData
		if l.propagate && l.Parent != nil {
			l.Parent.log(data)
		}
	}
}

// log creates record suitable for processing and sends it to messages chan.
func (l *Logger) log(data Record) {
	if _, ok := data[TimeKey]; !ok {
		data[TimeKey] = time.Now().UTC()
	}
	atomic.AddInt32(&l.toProcess, 1)
	l.rawRecords <- data
}

// wait blocks until all messages on messages channel are processed.
// Provided done channel will be closed when messages are processed to notify
// interested parties that they can unblock.
func (l *Logger) wait(done chan struct{}) {
	runtime.Gosched()
	l.notifyFinished <- done
}

// Wait block until all messages sent to logger are processed.
// If timeout is needed, see WaitTimeout.
func (l *Logger) Wait() {
	done := make(chan struct{})
	l.wait(done)
	<-done
}

// WaitTimeout blocks until all messages send to logger are processed or max
// specified amount of time.
// Boolean return value indicates if function returned because all messages
// were processed (true) or because timeout has expired (false).
func (l *Logger) WaitTimeout(t time.Duration) (ok bool) {
	done := make(chan struct{})
	timeout := time.After(t)
	l.wait(done)
	select {
	case <-done:
		return true
	case <-timeout:
		return false
	}
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
	d[EventKey] = event
	d[LevelKey] = level
	// make sure that number of items in data is even
	if len(pairs)%2 != 0 {
		pairs = append(pairs, "")
	}
	for i := 0; i < len(pairs); i += 2 {
		d[pairs[i]] = pairs[i+1]
	}
	l.log(d)
}

// LogRecord adds provided record to queue for processing.
func (l *Logger) LogRecord(record Record) {
	l.log(record)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Debug(event string, pairs ...string) {
	l.Log(DEBUG, event, pairs...)
}

// DebugRecord adds DEBUG level to provided record and queues it for processing.
func (l *Logger) DebugRecord(record Record) {
	record[LevelKey] = DEBUG
	l.LogRecord(record)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Info(event string, pairs ...string) {
	l.Log(INFO, event, pairs...)
}

// InfoRecord adds INFO level to provided record and queues it for processing.
func (l *Logger) InfoRecord(record Record) {
	record[LevelKey] = INFO
	l.LogRecord(record)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Warning(event string, pairs ...string) {
	l.Log(WARNING, event, pairs...)
}

// WarningRecord adds WARNING level to provided record and queues it for processing.
func (l *Logger) WarningRecord(record Record) {
	record[LevelKey] = WARNING
	l.LogRecord(record)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Error(event string, pairs ...string) {
	l.Log(ERROR, event, pairs...)
}

// ErrorRecord adds ERROR level to provided record and queues it for processing.
func (l *Logger) ErrorRecord(record Record) {
	record[LevelKey] = ERROR
	l.LogRecord(record)
}

// Critical creates log record and queues it for processing with CRITICAL level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Critical(event string, pairs ...string) {
	l.Log(CRITICAL, event, pairs...)
}

// CriticalRecord adds CRITICAL level to provided record and queues it for processing.
func (l *Logger) CriticalRecord(record Record) {
	record[LevelKey] = INFO
	l.LogRecord(record)
}
