# Ligno
Ligno is async structured logging library for golang.

Two main features that separate ligno from other logging libraries are that
all messages are logged asynchronously, which means that all expensive
operations related to logging (like writing to file or sending log message
over network) will not block execution of main application.

Other main feature is format of log record. Log record in ligno is just a
map holding key-value pairs. That's it. Because of this, records can be
formatted to any structured format (JSON, YAML, key=value) or to traditional
formats that we are used to seeing in log files.

## Example
Just a quick example of how usage might look like.
```go
package main

import "github.com/delicb/ligno"

func main() {
	// Create new logger with some context and handlers.
	l := ligno.New(
		ligno.Record{"always": "present"},
		[]ligno.Handler{
			&ligno.StdoutHandler{
				Formatter: &ligno.DefaultFormatter{},
			},
		},
	)
	// Log message. This will result in records in following format:
	// {
	//   "event": "Some event occurred",
	//   "key": "value",
	//   "key2": "value2",
	//   "always": "present",
	// }
	l.Info("Some event occurred.", "key", "value", "key2", "value2")
	// Since logger is async, wait for it to finish processing log messages
	l.Wait()
}
```

Output would look like:
```
2015-12-18 00:37:15.1895  INFO       Some event occurred. [always="present" key="value" key2="value2" ]
```

Setup of new still has some boilerplate code, I intend to spend some time on
figuring out better API for it.

## Note
At this point, ligno is just an idea that has been written down. But instead
of using pen and paper, I used editor and wrote it in code. So, expect breaking
changes and do not use it in production just yet - I am not (but I intend to).

## TODO
This is only just a skeleton, much is left.
- Implementation of other handlers (only stdout handler is available).
- Implementation of other formatters, like feature rich JSON formatter, YAML formatter, etc...
- Integration with other frameworks (logging, web or other).
- Documentation.
- Tests.
