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
	defaultLog.log(d)
}

// LogRecord adds provided record to queue for processing.
func LogRecord(record Record) {
	defaultLog.log(record)
}

// Debug creates log record and queues it for processing with DEBUG level.
// Additional parameters have same semantics as in Log method.
func Debug(event string, pairs ...string) {
	defaultLog.Log(DEBUG, event, pairs...)
}

// DebugRecord adds DEBUG level to provided record and queues it for processing.
func DebugRecord(record Record) {
	record[LevelKey] = DEBUG
	defaultLog.LogRecord(record)
}

// Info creates log record and queues it for processing with INFO level.
// Additional parameters have same semantics as in Log method.
func Info(event string, pairs ...string) {
	defaultLog.Log(INFO, event, pairs...)
}

// InfoRecord adds INFO level to provided record and queues it for processing.
func InfoRecord(record Record) {
	record[LevelKey] = INFO
	defaultLog.LogRecord(record)
}

// Warning creates log record and queues it for processing with WARNING level.
// Additional parameters have same semantics as in Log method.
func Warning(event string, pairs ...string) {
	defaultLog.Log(WARNING, event, pairs...)
}

// WarningRecord adds WARNING level to provided record and queues it for processing.
func WarningRecord(record Record) {
	record[LevelKey] = WARNING
	defaultLog.LogRecord(record)
}

// Error creates log record and queues it for processing with ERROR level.
// Additional parameters have same semantics as in Log method.
func Error(event string, pairs ...string) {
	defaultLog.Log(ERROR, event, pairs...)
}

// ErrorRecord adds ERROR level to provided record and queues it for processing.
func ErrorRecord(record Record) {
	record[LevelKey] = ERROR
	defaultLog.LogRecord(record)
}

// Critical creates log record and queues it for processing with CRITICAL level.
// Additional parameters have same semantics as in Log method.
func Critical(event string, pairs ...string) {
	defaultLog.Log(CRITICAL, event, pairs...)
}

// CriticalRecord adds CRITICAL level to provided record and queues it for processing.
func CriticalRecord(record Record) {
	record[LevelKey] = INFO
	defaultLog.LogRecord(record)
}
