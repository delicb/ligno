package ligno

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var rootLogger *Logger = createLogger(nil, nil, NOTSET, 2048, 2048)

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
	// Context in which logger is operating. Basically, this is set of
	// key-value pairs that will be added to every record logged with this
	// logger. They have lowest priority.
	context Ctx
	// handler is backed for processing records.
	handler struct {
		sync.RWMutex
		val Handler
	}

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
		children []*Logger
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
	level struct {
		sync.RWMutex
		val Level
	}
}

// New creates new instance of logger and starts it so it is ready for processing.
func New(context Ctx, handler Handler) *Logger {
	return rootLogger.SubLogger(context, handler, false)
}

func createLogger(context Ctx, handler Handler, level Level, recordsBuffer uint, rawRecordsBuffer uint) *Logger {
	l := &Logger{
		context:        context,
		records:        make(chan Record, recordsBuffer),
		rawRecords:     make(chan Record, rawRecordsBuffer),
		notifyFinished: make(chan chan struct{}),
	}
	// no need to lock access to handler, state or level here
	// since we just created logger and nobody can use it anywhere else.
	l.handler.val = handler
	l.state.val = loggerRunning
	l.level.val = level
	// start logger pipeline
	go l.handle()
	go l.processRecords()
	return l
}

// SubLogger creates and returns new logger whose parent is current logger.
func (l *Logger) SubLogger(context Ctx, handler Handler, preventPropagation bool) *Logger {
	newLogger := createLogger(context, handler, DEBUG, 2048, 2048)
	l.addChild(newLogger, preventPropagation)
	return newLogger
}

func (l *Logger) addChild(child *Logger, preventPropagation bool) {
	l.relationship.Lock()
	l.relationship.children = append(l.relationship.children, child)
	l.relationship.Unlock()

	child.relationship.Lock()
	child.relationship.parent = l
	child.relationship.preventPropagation = preventPropagation
	child.relationship.Unlock()
}

func (l *Logger) removeChild(child *Logger) {
	if l == nil {
		fmt.Println("Got nil receiver")
	}
	l.relationship.Lock()
	defer l.relationship.Unlock()
	var removeIndex int = -1
	for i, ch := range l.relationship.children {
		if child == ch {
			removeIndex = i
			break
		}
	}
	// remove it only if we found it
	if removeIndex > -1 {
		copy(l.relationship.children[removeIndex:], l.relationship.children[removeIndex+1:])
		// if we do not do this there is potential memory leak, since we are holding pointers
		l.relationship.children[len(l.relationship.children)-1] = nil
		l.relationship.children = l.relationship.children[:len(l.relationship.children)-1]
	}
}

// handle is log record processor which takes records from chan and invokes all handlers.
func (l *Logger) handle() {
	var handler Handler
	l.handler.RLock()
	if l.handler.val != nil {
		handler = l.handler.val
	} else {
		handler = stdoutHandler
	}
	l.handler.RUnlock()

	var notifyFinished chan struct{}
	for {
		select {
		case record := <-l.records:
			handler.Handle(record)
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
	if l.state.val == loggerStopped || !l.shouldProcessLevel(record.Level) {
		return
	}

	atomic.AddInt32(&l.toProcess, 1)
	l.rawRecords <- record
}

// Stop stops listening for new messages sent to this logger.
// Messages already sent will be processed, but all new messages will
// silently be dropped.
// Stopping loggers stops processing goroutines and cleans up resources.
func (l *Logger) Stop() {
	l.state.Lock()
	defer l.state.Unlock()
	l.state.val = loggerStopped
	close(l.rawRecords)
	if l.relationship.parent != nil {
		l.relationship.parent.removeChild(l)
	}
}

// StopAndWait stops listening for new messages sent to this logger and
// blocks until all previously arrived messages are processed.
// Records already sent will be processed, but all new messages will
// silently be dropped.
func (l *Logger) StopAndWait() {
	l.Stop()
	l.Wait()
}

// StopAndWaitTimeout stops listening for new messages sent to this logger and
// blocks until all previously sent message are processed or max provided duration.
// Records already sent will be processed, but all new messages will
// silently be dropped. Return value indicates if all messages are processed (true)
// or if provided timeout expired (false)
func (l *Logger) StopAndWaitTimeout(t time.Duration) (finished bool) {
	l.Stop()
	return l.WaitTimeout(t)
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
	var d = make(Ctx)

	//	d[EventKey] = event
	//	d[LevelKey] = level
	//	d[TimeKey] = creationTime
	// make sure that number of items in data is even
	if len(pairs)%2 != 0 {
		pairs = append(pairs, "")
	}
	for i := 0; i < len(pairs); i += 2 {
		d[pairs[i]] = pairs[i+1]
	}
	r := Record{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
		Context: d,
	}
	l.log(r)
}

// LogCtx adds provided message in specified level.
func (l *Logger) LogCtx(level Level, message string, data Ctx) {
	r := Record{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
		Context: data,
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

func (l *Logger) shouldProcessLevel(level Level) bool {
	l.level.RLock()
	defer l.level.RUnlock()
	return l.level.val <= level
}

// IsDebug returns true if logger will process messages in DEBUG level
func (l *Logger) IsDebug() bool {
	return l.shouldProcessLevel(DEBUG)
}

// IsInfo returns true if logger will process messages in INFO level
func (l *Logger) IsInfo() bool {
	return l.shouldProcessLevel(DEBUG)
}

// IsWarning returns true if logger will process messages in WARNING level
func (l *Logger) IsWarning() bool {
	return l.shouldProcessLevel(DEBUG)
}

// IsError returns true if logger will process messages in ERROR level
func (l *Logger) IsError() bool {
	return l.shouldProcessLevel(DEBUG)
}

// IsCritical returns true if logger will process messages in Critical level
func (l *Logger) IsCritical() bool {
	return l.shouldProcessLevel(DEBUG)
}

// IsLevel return true if logger will process messages in provided level.
func (l *Logger) IsLevel(level Level) bool {
	return l.shouldProcessLevel(level)
}
