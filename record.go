package ligno

import "time"

// Record is description of one event that occurred and that needs to be logged.
// It is key-value store with strings as keys.
type Record map[string]interface{}

// Constants for keys in log records that are library treats specially.
const (
	TIME  = "time"
	LEVEL = "level"
	EVENT = "event"
)

const defaultLevel = INFO

// Level returns log level for record.
// If, by some chance, level is not part of record, default level (INFO) is returned.
func (d Record) Level() Level {
	if l, ok := d[LEVEL]; ok {
		if ll, ok := l.(Level); ok {
			return ll
		}
		return defaultLevel
	}
	return defaultLevel
}

// Time returns time when log record was created.
// If, by some chance, time is not part of record, empty value for time.Time is returned.
func (d Record) Time() time.Time {
	if t, ok := d[TIME]; ok {
		if tt, ok := t.(time.Time); ok {
			return tt
		}
		return time.Time{}
	}
	return time.Time{}
}

// Event returns event message for log record.
// If, by some chance, event is not part of record, empty string will be returned.
func (d Record) Event() string {
	if e, ok := d[EVENT]; ok {
		if ee, ok := e.(string); ok {
			return ee
		}
		return ""
	}
	return ""
}
