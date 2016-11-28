// +build !windows,!nacl,!plan9

package ligno

import "log/syslog"

// syslogHandler sends all messages to local syslog server.
type syslogHandler struct {
	Formatter Formatter
	Tag       string
	Priority  syslog.Priority
	writer    *syslog.Writer
}

// SyslogHandler creates new syslog handler with provided config variables.
func SyslogHandler(formatter Formatter, tag string, priority syslog.Priority) Handler {
	writer, err := syslog.New(syslog.LOG_DEBUG, tag)
	if err != nil {
		panic(err)
	}
	return &syslogHandler{
		Formatter: formatter,
		Tag:       tag,
		Priority:  priority,
		writer:    writer,
	}
}

// Handle passes all messages to syslog server. Message priorities are
// translated to syslog compatible priorities.
func (sh *syslogHandler) Handle(record Record) error {

	msg := string(sh.Formatter.Format(record))
	switch record.Level {
	case NOTSET:
		return sh.writer.Info(msg)
	case DEBUG:
		return sh.writer.Debug(msg)
	case INFO:
		return sh.writer.Info(msg)
	case WARNING:
		return sh.writer.Warning(msg)
	case ERROR:
		return sh.writer.Err(msg)
	case CRITICAL:
		return sh.writer.Crit(msg)
	default:
		return sh.writer.Info(msg)
	}
}

// Close closes connection with syslog server.
func (sh *syslogHandler) Close() {
	sh.writer.Close()
}
