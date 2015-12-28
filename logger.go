package ligno

import (
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"strings"
)

// root logger is parent of all loggers and it always exists.
var rootLogger = createLogger("", LoggerOptions{
	Context:    nil,
	Handler:    StreamHandler(os.Stderr, SimpleFormat()),
	BufferSize: 2048,
})

// WaitAll blocks until all loggers are finished with message processing.
func WaitAll() {
	rootLogger.Wait()
}

// WaitAllTimeout blocks until all messages send to all loggers are processed or max
// specified amount of time.
// Boolean return value indicates if function returned because all messages
// were processed (true) or because timeout has expired (false).
func WaitAllTimeout(t time.Duration) bool {
	return rootLogger.WaitTimeout(t)
}

type loggerState uint8

const (
	loggerRunning loggerState = iota
	loggerStopped
)

// Logger is central data type in ligno which represents logger itself.
// Logger is first level of processing events. It creates them and
// queues for async processing. It holds slice of Handlers that process
// messages and context (set of key-value pairs that will be include
// in every log record).
type Logger struct {
	// name is name of this logger.
	name string
	// Context in which logger is operating. Basically, this is set of
	// key-value pairs that will be added to every record logged with this
	// logger. They have lowest priority.
	context Ctx
	// handler is backed for processing records.
	handler *replaceableHandler
	// handlerChanged is notification mechanism to notify working goroutines
	// that they need to switch their handler for further processing.
	handlerChanged chan Handler

	// relationship holds information about logger parent and children.
	relationship struct {
		sync.RWMutex
		// parent is this logger's parent. Final context of record is created
		// by combining all parents contexts and if preventPropagation is false
		// all records will be sent to parent to processing as well.
		// parent should be immutable once it is set, so it should be safe to
		// read it without concurrency protection.
		parent *Logger
		// children is slice of all loggers that have this logger as parent.
		//		children []*Logger
		children map[string]*Logger
		// preventPropagation is flag that indicates if propagation of log
		// records to parent should be prevented.
		preventPropagation bool
	}
	// rawMessages is channel for queueing and buffering raw messages from
	// application which needs to be merged with context and submitted
	// to final processing
	rawRecords chan Record
	// records is channel for queueing and buffering log records.
	records chan Record
	// notifyFinished is channel of channels. When someone wants to be notified
	// when logger processed all queued records, it sends channel that will be
	// closed after last queued record is processed to notifyFinished.
	notifyFinished chan chan struct{}
	// toProcess is number of messages left to process in this logger.
	toProcess int32
	// state represents state in which logger is currently
	state struct {
		sync.RWMutex
		val loggerState
	}
	// level is lowest level that this logger will process
	level Level
}

// LoggerOptions is container for configuration options for logger instances.
// Empty value is valid for initializing logger.
type LoggerOptions struct {
	// Context that logger should have.
	Context Ctx
	// Handler for processing records.
	Handler Handler
	// Level is minimal level that logger will process.
	Level Level
	// BufferSize is size of buffer for records that will be process async.
	BufferSize int
	// PreventPropagation is flag that indicates if records should be passed
	// to parent logger for processing.
	PreventPropagation bool
}

// createLogger creates new instance of logger, initializes all values based
// on provided options and starts worker goroutines.
func createLogger(name string, options LoggerOptions) *Logger {
	rh := new(replaceableHandler)
	rh.Replace(options.Handler)
	var buffSize = 0
	if options.BufferSize > 0 {
		buffSize = options.BufferSize
	} else {
		buffSize = 1024
	}
	l := &Logger{
		name:           name,
		context:        options.Context,
		records:        make(chan Record, buffSize),
		rawRecords:     make(chan Record, buffSize),
		notifyFinished: make(chan chan struct{}),
		handler:        rh,
		level:          options.Level,
	}
	// no need to lock access to state here since we just created logger
	// and nobody can use it anywhere else at the moment.
	l.state.val = loggerRunning
	l.relationship.children = make(map[string]*Logger)
	l.relationship.preventPropagation = options.PreventPropagation
	go l.handle()
	go l.processRecords()
	return l
}

// SubLogger creates new logger that has current logger as parent with default
// options and starts it so it is ready for message processing.
func (l *Logger) SubLogger(name string) *Logger {
	newLogger := createLogger(name, LoggerOptions{})
	l.addChild(newLogger)
	return newLogger
}

// SubLoggerOptions creates new logger that has current logger as parent with
// provided options and starts it so it is ready for message processing.
func (l *Logger) SubLoggerOptions(name string, options LoggerOptions) *Logger {
	newLogger := createLogger(name, options)
	l.addChild(newLogger)
	return newLogger
}

// GetLogger returns logger with provided name (creating it if needed).
// Name is dot-separated string with parent logger names and this function
// will create all intermediate loggers with default options.
func GetLogger(name string) *Logger {
	current := rootLogger
	for _, part := range strings.Split(name, ".") {
		current.relationship.RLock()
		child, ok := current.relationship.children[part]
		current.relationship.RUnlock()
		if ok {
			current = child
		} else {
			current = current.SubLogger(part)
		}
	}
	return current
}

// GetLoggerOptions returns logger with provided name (creating it if needed).
// Name is dot-separated string with parent loggers and this function will
// create all intermediate loggers with default options. Provided options will
// be applied only to last logger in chain. If all loggers in chain already
// exist, no new loggers will be creates and provided options will be discarded.
// If options different then default are needed for intermediate loggers,
// create them first with appropriate options.
func GetLoggerOptions(name string, options LoggerOptions) *Logger {
	current := rootLogger
	parts := strings.Split(name, ".")
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		current.relationship.RLock()
		child, ok := current.relationship.children[part]
		current.relationship.RUnlock()
		if ok {
			current = child
		} else {
			if i == len(parts) - 1 {
				current = current.SubLoggerOptions(part, options)
			} else {
				current = current.SubLogger(part)
			}
		}
	}
	return current
}

func (l *Logger) addChild(child *Logger) {
	l.relationship.Lock()
	//	l.relationship.children = append(l.relationship.children, child)
	l.relationship.children[child.name] = child
	l.relationship.Unlock()

	child.relationship.Lock()
	child.relationship.parent = l
	child.relationship.Unlock()
}

func (l *Logger) removeChild(child *Logger) {
	l.relationship.Lock()
	defer l.relationship.Unlock()
	delete(l.relationship.children, child.name)
}

// SetHandler set handler to this logger to be used from now on.
func (l *Logger) SetHandler(handler Handler) {
	l.handler.Replace(handler)
}

// Handler returns current handler for this logger
func (l *Logger) Handler() Handler {
	return l.handler.Handler()
}

// Level returns minimal level that this logger will process.
func (l *Logger) Level() Level {
	return l.level
}

// Name returns name of this logger.
func (l *Logger) Name() string {
	return l.name
}

// FullName returns name of this logger prefixed with name of its parent,
// separated by ".". This happens recursively, so return value will contain
// names of all parents.
func (l *Logger) FullName() string {
	if l.relationship.parent != nil {
		parentFullName := l.relationship.parent.FullName()
		if parentFullName != "" {
			return l.relationship.parent.FullName() + "." + l.name
		}
	}
	return l.name
}

// handle is log record processor which takes records from chan and invokes all handlers.
func (l *Logger) handle() {
	var notifyFinished chan struct{}
	for {
		select {
		case record := <-l.records:
			l.handler.Handle(record)

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
	for {
		select {
		case record, ok := <-l.rawRecords:
			if !ok {
				return
			}
			//	for data := range l.rawRecords {
			var current = l
			var loggerChain = make([]*Logger, 0, 5)

			// create list of all parents
			for current != nil {
				loggerChain = append(loggerChain, current)
				current = current.relationship.parent
			}
			// merge context of all parents backwards, because children can override parents
			mergedData := make(Ctx)
			for i := len(loggerChain) - 1; i >= 0; i-- {
				for k, v := range loggerChain[i].context {
					mergedData[k] = v
				}
			}
			// finally, add provided data to override all context keys
			for k, v := range record.Context {
				mergedData[k] = v
			}
			record.Context = mergedData
			l.records <- record
			if !l.relationship.preventPropagation && l.relationship.parent != nil {
				l.relationship.parent.log(record)
			}
		}
	}
}

// log creates record suitable for processing and sends it to messages chan.
func (l *Logger) log(record Record) {
	l.state.RLock()
	defer l.state.RUnlock()
	if l.state.val == loggerStopped || !l.IsEnabledFor(record.Level) {
		return
	}
	atomic.AddInt32(&l.toProcess, 1)
	l.rawRecords <- record
}

// Stop stops listening for new messages sent to this logger.
// Messages already sent will be processed, but all new messages will
// silently be dropped.
// Stopping loggers stops processing goroutines and cleans up resources.
func (l *Logger) stopAndWait(waitFunc func()) {
	l.state.Lock()
	defer l.state.Unlock()
	l.state.val = loggerStopped
	close(l.rawRecords)
	if l.relationship.parent != nil {
		l.relationship.parent.removeChild(l)
	}
	waitFunc()
	if handlerCloser, ok := l.Handler().(HandlerCloser); ok {
		handlerCloser.Close()
	}
}

// StopAndWait stops listening for new messages sent to this logger and
// blocks until all previously arrived messages are processed.
// Records already sent will be processed, but all new messages will
// silently be dropped.
func (l *Logger) StopAndWait() {
	l.stopAndWait(l.Wait)
}

// StopAndWaitTimeout stops listening for new messages sent to this logger and
// blocks until all previously sent message are processed or max provided duration.
// Records already sent will be processed, but all new messages will
// silently be dropped. Return value indicates if all messages are processed (true)
// or if provided timeout expired (false)
func (l *Logger) StopAndWaitTimeout(t time.Duration) (finished bool) {
	finished = false
	l.stopAndWait(func() {
		finished = l.WaitTimeout(t)
	})
	return finished
}

// IsRunning returns boolean indicating if this logger is still running.
func (l *Logger) IsRunning() bool {
	l.state.RLock()
	defer l.state.RUnlock()
	return l.state.val == loggerRunning
}

// wait blocks until all messages on messages channel are processed.
// Provided done channel will be closed when messages are processed to notify
// interested parties that they can unblock.
func (l *Logger) wait(done chan struct{}) {
	runtime.Gosched()
	l.relationship.RLock()
	defer l.relationship.RUnlock()
	var wg sync.WaitGroup
	wg.Add(len(l.relationship.children) + 1)
	go func() {
		l.notifyFinished <- done
		wg.Done()
	}()
	for _, child := range l.relationship.children {
		go func(l *Logger) {
			chDone := make(chan struct{})
			l.wait(chDone)
			<-chDone
			wg.Done()
		}(child)
	}
	wg.Wait()
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
func (l *Logger) WaitTimeout(t time.Duration) (finished bool) {
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
func (l *Logger) Log(level Level, message string, pairs ...string) {
	// if level is not sufficient, do not proceed to avoid unneeded allocations
	if !l.IsEnabledFor(level) {
		return
	}
	var d = make(Ctx)

	// make sure that number of items in data is even
	if len(pairs)%2 != 0 {
		pairs = append(pairs, []string{"", "error", "missing key"}...)
	}
	for i := 0; i < len(pairs); i += 2 {
		d[pairs[i]] = pairs[i+1]
	}

	r := Record{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
		Context: d,
		Logger:  l,
	}
	l.log(r)
}

// LogCtx adds provided message in specified level.
func (l *Logger) LogCtx(level Level, message string, data Ctx) {
	// if level is not sufficient, do not proceed to avoid unneeded allocations
	if !l.IsEnabledFor(level) {
		return
	}

	r := Record{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
		Context: data,
		Logger:  l,
	}
	l.log(r)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Debug(message string, pairs ...string) {
	l.Log(DEBUG, message, pairs...)
}

// DebugCtx logs message in DEBUG level with provided context.
func (l *Logger) DebugCtx(message string, ctx Ctx) {
	l.LogCtx(DEBUG, message, ctx)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Info(message string, pairs ...string) {
	l.Log(INFO, message, pairs...)
}

// InfoCtx logs message in INFO level with provided context.
func (l *Logger) InfoCtx(message string, ctx Ctx) {
	l.LogCtx(INFO, message, ctx)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Warning(message string, pairs ...string) {
	l.Log(WARNING, message, pairs...)
}

// WarningCtx logs message in WARNING level with provided context.
func (l *Logger) WarningCtx(message string, ctx Ctx) {
	l.LogCtx(WARNING, message, ctx)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Error(message string, pairs ...string) {
	l.Log(ERROR, message, pairs...)
}

// ErrorCtx logs message in ERROR level with provided context.
func (l *Logger) ErrorCtx(message string, ctx Ctx) {
	l.LogCtx(ERROR, message, ctx)
}

// Critical creates log record and queues it for processing with CRITICAL level.
// Additional parameters have same semantics as in Log method.
func (l *Logger) Critical(message string, pairs ...string) {
	l.Log(CRITICAL, message, pairs...)
}

// CriticalCtx logs message in CRITICAL level with provided context.
func (l *Logger) CriticalCtx(message string, ctx Ctx) {
	l.LogCtx(CRITICAL, message, ctx)
}

// IsEnabledFor returns true if logger will process records with provided level.
func (l *Logger) IsEnabledFor(level Level) bool {
	return l.Level() <= level
}

// IsDebug returns true if logger will process messages in DEBUG level
func (l *Logger) IsDebug() bool {
	return l.IsEnabledFor(DEBUG)
}

// IsInfo returns true if logger will process messages in INFO level
func (l *Logger) IsInfo() bool {
	return l.IsEnabledFor(DEBUG)
}

// IsWarning returns true if logger will process messages in WARNING level
func (l *Logger) IsWarning() bool {
	return l.IsEnabledFor(DEBUG)
}

// IsError returns true if logger will process messages in ERROR level
func (l *Logger) IsError() bool {
	return l.IsEnabledFor(DEBUG)
}

// IsCritical returns true if logger will process messages in Critical level
func (l *Logger) IsCritical() bool {
	return l.IsEnabledFor(DEBUG)
}

// IsLevel return true if logger will process messages in provided level.
func (l *Logger) IsLevel(level Level) bool {
	return l.IsEnabledFor(level)
}
