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
