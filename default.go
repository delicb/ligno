package ligno

import (
	"fmt"
	"os"
)

// Log creates record and queues it for processing.
// Required parameters are level for record and event that occurred. Any
// additional parameters will be transformed to key-value pairs for record
// in order in which they were provided. There should be even number of them,
// but in case that there is on number of parameters, empty string is
// appended. Example:
//   defaultLog.Log(INFO, "User logged in", "user_id", user_id, "platform", PLATFORM_NAME)
// will be translated into log record with following keys:
//  {LEVEL: INFO", EVENT: "User logged in", "user_id": user_id, "platform": PLATFORM_NAME}
func Log(level Level, event string, pairs ...interface{}) {
	rootLogger.Log(2, level, event, pairs...)
}

// LogCtx adds provided message in specified level.
func LogCtx(level Level, message string, ctx Ctx) {
	rootLogger.LogCtx(2, level, message, ctx)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func Debug(event string, pairs ...interface{}) {
	rootLogger.Log(2, DEBUG, event, pairs...)
}

// DebugCtx logs message in DEBUG level with provided context.
func DebugCtx(message string, ctx Ctx) {
	rootLogger.LogCtx(2, DEBUG, message, ctx)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func Info(event string, pairs ...interface{}) {
	rootLogger.Log(2, INFO, event, pairs...)
}

// InfoCtx logs message in INFO level with provided context.
func InfoCtx(message string, ctx Ctx) {
	rootLogger.LogCtx(2, INFO, message, ctx)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func Warning(event string, pairs ...interface{}) {
	rootLogger.Log(2, WARNING, event, pairs...)
}

// WarningCtx logs message in WARNING level with provided context.
func WarningCtx(message string, ctx Ctx) {
	rootLogger.LogCtx(2, WARNING, message, ctx)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func Error(event string, pairs ...interface{}) {
	rootLogger.Log(2, ERROR, event, pairs...)
}

// ErrorCtx logs message in ERROR level with provided context.
func ErrorCtx(message string, ctx Ctx) {
	rootLogger.LogCtx(2, ERROR, message, ctx)
}

// Critical creates log record and queues it for processing with CRITICAL level.
// Additional parameters have same semantics as in Log method.
func Critical(event string, pairs ...interface{}) {
	rootLogger.Log(2, CRITICAL, event, pairs...)
}

// CriticalCtx logs message in CRITICAL level with provided context.
func CriticalCtx(message string, ctx Ctx) {
	rootLogger.LogCtx(2, CRITICAL, message, ctx)
}

// Printf formats message according to stdlib rules and logs it in INFO level.
func Printf(format string, v ...interface{}) {
	rootLogger.Log(2, INFO, fmt.Sprintf(format, v...))
}

// Print formats message according to stdlib rules and logs it in INFO level.
func Print(v ...interface{}) {
	rootLogger.Log(2, INFO, fmt.Sprint(v...))
}

// Println formats message according to stdlib rules and logs it in INFO level.
func Println(v ...interface{}) {
	rootLogger.Log(2, INFO, fmt.Sprintln(v...))
}

// Fatal formats message according to stdlib rules, logs it in CRITICAL level
// and exists application.
func Fatal(v ...interface{}) {
	rootLogger.Log(2, CRITICAL, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf formats message according to stdlib rules, logs it in CRITICAL level
// and exists application.
func Fatalf(format string, v ...interface{}) {
	rootLogger.Log(2, CRITICAL, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln formats message according to stdlib rules, logs it in CRITICAL level
// and exists application.
func Fatalln(v ...interface{}) {
	rootLogger.Log(2, CRITICAL, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic formats message according to stdlib rules, logs it in CRITICAL level
// and panics.
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	rootLogger.Log(2, CRITICAL, s)
	panic(s)
}

// Panicf formats message according to stdlib rules, logs it in CRITICAL level
// and panics.
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	rootLogger.Log(2, CRITICAL, s)
	panic(s)
}

// Panicln formats message according to stdlib rules, logs it in CRITICAL level
// and panics.
func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	rootLogger.Log(2, CRITICAL, s)
	panic(s)
}

// SetHandler sets new handler for default logger.
func SetHandler(handler Handler) {
	rootLogger.SetHandler(handler)
}
