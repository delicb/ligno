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
	rootLogger.Log(level, event, pairs...)
}

// LogCtx adds provided message in specified level.
func LogCtx(level Level, message string, ctx Ctx) {
	rootLogger.LogCtx(level, message, ctx)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func Debug(event string, pairs ...interface{}) {
	rootLogger.Debug(event, pairs...)
}

// DebugCtx logs message in DEBUG level with provided context.
func DebugCtx(message string, ctx Ctx) {
	rootLogger.DebugCtx(message, ctx)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func Info(event string, pairs ...interface{}) {
	rootLogger.Info(event, pairs...)
}

// InfoCtx logs message in INFO level with provided context.
func InfoCtx(message string, ctx Ctx) {
	rootLogger.InfoCtx(message, ctx)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func Warning(event string, pairs ...interface{}) {
	rootLogger.Warning(event, pairs...)
}

// WarningCtx logs message in WARNING level with provided context.
func WarningCtx(message string, ctx Ctx) {
	rootLogger.WarningCtx(message, ctx)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func Error(event string, pairs ...interface{}) {
	rootLogger.Error(event, pairs...)
}

// ErrorCtx logs message in ERROR level with provided context.
func ErrorCtx(message string, ctx Ctx) {
	rootLogger.ErrorCtx(message, ctx)
}

// Critical creates log record and queues it for processing with CRITICAL level.
// Additional parameters have same semantics as in Log method.
func Critical(event string, pairs ...interface{}) {
	rootLogger.Critical(event, pairs...)
}

// CriticalCtx logs message in CRITICAL level with provided context.
func CriticalCtx(message string, ctx Ctx) {
	rootLogger.CriticalCtx(message, ctx)
}

// SetHandler sets new handler for default logger.
func SetHandler(handler Handler) {
	rootLogger.SetHandler(handler)
}
