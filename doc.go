// Package ligno is async structured logging library for golang.
//
// Two main features that separate ligno from other logging libraries are that
// all messages are logged asynchronously, which means that all expensive
// operations related to logging (like writing to file or sending log message
// over network) will not block execution of main application.
//
// Other main feature is format of log record. Log record in ligno is just a
// map holding key-value pairs. That's it. Because of this, records can be
// formatted to any structured format (JSON, YAML, key=value) or to traditional
// formats that we are used to seeing in log files.
package ligno // import "go.delic.rs/ligno"
