package ligno

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Formatter is interface for converting log record to string representation.
type Formatter interface {
	Format(record Record) string
}

// DefaultFormatter converts log record to simple string for printing.
type DefaultFormatter struct{
	TimeFormat string
}

// defaultTimeFormat is formatting string for time for DefaultFormatter
const defaultTimeFormat = "2006-01-02 15:05:06.0000"

// Format converts provided log record to format suitable for printing in one line.
// String produced resembles traditional log message.
func (df *DefaultFormatter) Format(record Record) string {
	var timeFormat = defaultTimeFormat
	if df.TimeFormat != "" {
		timeFormat = df.TimeFormat
	}
	time := record.Time.Format(timeFormat)
	var buff bytes.Buffer

	ctx := record.Context
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < len(keys); i++ {
		k := keys[i]
		v := strconv.Quote(fmt.Sprintf("%+v", ctx[k]))
		if strings.IndexFunc(k, needsQuote) >= 0 || k == "" {
			k = strconv.Quote(k)
		}
		buff.WriteString(fmt.Sprintf("%s=%+v", k, v))
		if i < len(keys)-1 {
			buff.WriteString(" ")
		}
	}
	return fmt.Sprintf("%-25s %-10s %-15s [%s]\n", time, record.Level, record.Message, buff.String())
}

// defaultFormatter is instance of DefaultFormatter.
var defaultFormatter = &DefaultFormatter{}

// Needs quote determines if provided rune is such that word that contains this
// rune needs to be quoted.
func needsQuote(r rune) bool {
	return r == ' ' || r == '"' || r == '\\' || r == '=' ||
		!unicode.IsPrint(r)
}

// JSONFormatter is simple formatter that only marshals log record to json.
type JSONFormatter struct{
	Indent bool
}

// Format returns JSON representation of provided record.
func (jf *JSONFormatter) Format(record Record) string {
	var marshaled []byte
	var err error
	if jf.Indent {
		marshaled, err = json.MarshalIndent(record.Context, "", "    ")
	} else {
		marshaled, err = json.Marshal(record)
	}
	if err != nil {
		panic(err)
	}
	return string(marshaled)
}
