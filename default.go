package ligno

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

// SetHandler sets new handler for default logger.
func SetHandler(handler Handler) {
	rootLogger.SetHandler(handler)
}
