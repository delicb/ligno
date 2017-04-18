# Ligno
[![Go Report Card](https://goreportcard.com/badge/github.com/delicb/ligno)](https://goreportcard.com/report/github.com/delicb/ligno)
[![Build Status](https://travis-ci.org/delicb/ligno.svg?branch=master)](https://travis-ci.org/delicb/ligno)
[![codecov](https://codecov.io/gh/delicb/ligno/branch/master/graph/badge.svg)](https://codecov.io/gh/delicb/ligno)
![status](https://img.shields.io/badge/status-beta-red.svg)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/delicb/ligno)

Ligno is async structured logging library for golang.

Two main features that separate ligno from other logging libraries are that
all messages are logged asynchronously, which means that all expensive
operations related to logging (like writing to file or sending log message
over network) will not block execution of main application.

Other main feature is format of log record. Log record in ligno is just a
map holding key-value pairs. That's it. Because of this, records can be
formatted to any structured format (JSON, YAML, key=value) or to traditional
formats that we are used to seeing in log files.

## Install
Run `go get github.com/delicb/ligno` from command line.

## Example
Just a quick example of how usage might look like.
```go
package main

import (
	"os"

	"github.com/delicb/ligno"
)

func main() {
	// Create new logger with some context and handlers.
	l := ligno.GetLoggerOptions("myLogger", ligno.LoggerOptions{
		Context:            ligno.Ctx{"always": "present"},
		Handler:            ligno.StreamHandler(os.Stdout, ligno.JSONFormat(true)),
		Level:              ligno.INFO,
		BufferSize:         256,
		PreventPropagation: true,
	})

	// Log message. This will result in records in following format:
	//	{
	//    "time": "2016-01-07T01:06:10.937122038Z",
	//    "level": "INFO",
	//    "message": "Some event occurred.",
	//    "context": {
	//        "always": "present",
	//        "key": "value",
	//        "key2": "value2"
	//    }
	//	}
	l.Info("Some event occurred.", "key", "value", "key2", "value2")

	// Log message in DEBUG level, which will be ignored since this logger is
	// configured to log only INFO messages and above
	l.Debug("Will be discarded.")

	// Since logger is async, wait for it to finish processing log messages
	l.Wait()
}

```

Output would look like:
```
{
    "time": "2016-02-09T20:58:39.313122319Z",
    "level": "INFO",
    "message": "Some event occurred.",
    "context": {
        "always": "present",
        "key": "value",
        "key2": "value2"
    }
}
```

Setup of new logger still has some boilerplate code, I intend to spend some time on
figuring out better API for it.

## Benchmarks
I have not used builtin golang benchmarks to measure performance yet, but I did hack up small script
that compares ligno with bunch of other logging frameworks, including golang stdlib. With every logger 
some number of messages are logged to stdout (1024 by default) and two numbers are produced for every logger: 
average time and total time. Average time is time spent in logging library (overhead that application using
logger sees) for one message. Total is time that logging library took to process all messages. Since ligno is
async, this is not simply *average* * *number* of messages, it will be higher, so I included it to compare
overall performance with other logging libraries. For me, time spent in logging library for single message is
much more important then total time spent processing log messages (since that overhead will come at the end of
program execution), but I think that it is only fair to include both.

Benchmarking script can be found in benchmark folder, but this is example of one output:
```
Logging 1024 messages.
Ligno           average time:      697ns, total time:     49.149567ms
resenje-logging average time:      824ns, total time:     78.261047ms
stdlib          average time:   39.565µs, total time:     40.569742ms
gommon          average time:   42.024µs, total time:     43.096594ms
seelog          average time:   63.808µs, total time:     66.869716ms
Log15           average time:   65.806µs, total time:     67.447484ms
logrus          average time:   67.856µs, total time:     69.542356ms
logxi           average time:   76.046µs, total time:     77.920668ms
Total execution time: 78.463182ms
```

## Credits
I was reading bunch of articles and source code for existing logging libraries so if you
recognize some pattern from somewhere else, it is quite possible that I have seen it there.

## Note
At this point, ligno is just an idea that has been written down. But instead
of using pen and paper, I used editor and wrote it in code. So, expect breaking
changes and do not use it in production just yet - I am not (but I intend to).

## TODO
This is only just a skeleton, much is left.
- Implementation of other handlers.
- Implementation of other formatters, like feature rich JSON formatter, YAML formatter, etc...
- Integration with other frameworks (logging, web or other).
- Documentation.
- Tests.
