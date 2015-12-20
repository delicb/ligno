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
func Log(level Level, event string, pairs ...string) {
	defaultLog.Log(level, event, pairs...)
}

// LogRecord adds provided record to queue for processing.
func LogRecord(record Record) {
	defaultLog.LogRecord(record)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func Debug(event string, pairs ...string) {
	defaultLog.Debug(event, pairs...)
}

// DebugRecord adds DEBUG level to provided record and queues it for processing.
func DebugRecord(record Record) {
	defaultLog.DebugRecord(record)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func Info(event string, pairs ...string) {
	defaultLog.Info(event, pairs...)
}

// InfoRecord adds INFO level to provided record and queues it for processing.
func InfoRecord(record Record) {
	defaultLog.InfoRecord(record)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func Warning(event string, pairs ...string) {
	defaultLog.Warning(event, pairs...)
}

// WarningRecord adds WARNING level to provided record and queues it for processing.
func WarningRecord(record Record) {
	defaultLog.WarningRecord(record)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func Error(event string, pairs ...string) {
	defaultLog.Error(event, pairs...)
}

// ErrorRecord adds ERROR level to provided record and queues it for processing.
func ErrorRecord(record Record) {
	defaultLog.ErrorRecord(record)
}

// Critical creates log record and queues it for processing with CRITICAL level.
// Additional parameters have same semantics as in Log method.
func Critical(event string, pairs ...string) {
	defaultLog.Critical(event, pairs...)
}

// CriticalRecord adds CRITICAL level to provided record and queues it for processing.
func CriticalRecord(record Record) {
	defaultLog.CriticalRecord(record)
}
