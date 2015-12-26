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
	Format(record Record) []byte
}

// FormatterFunc is function type that implements Formatter interface.
type FormatterFunc func (Record) []byte

// Format is implementation of Formatter interface. It just calls function.
func (ff FormatterFunc) Format(record Record) []byte {
	return ff(record)
}

// DefaultTimeFormat is default time format.
const DefaultTimeFormat = "2006-01-02 15:05:06.0000"

// SimpleFormat returns formatter that formats record with bare minimum of information.
// Intention of this formatter is to simulate standard library formatter.
func SimpleFormat() Formatter {
	return FormatterFunc(func(record Record) []byte {
		buff := new(bytes.Buffer)
		buff.WriteString(record.Time.Format(DefaultTimeFormat))
		buff.WriteString(" ")
		buff.WriteString(record.Message)
		buff.WriteString("\n")
		return buff.Bytes()
	})
}

// TerminalFormat returns formatter that produces records formatted for easy
// reading in terminal, but that are a bit richer then SimpleFormat (this one
// includes context keys)
func TerminalFormat() Formatter {
	return FormatterFunc(func(record Record) []byte {
		time := record.Time.Format(DefaultTimeFormat)
		buff := &bytes.Buffer{}

		fmt.Fprintf(buff, "%-25s %-10s %-15s", time, record.Level, record.Message)

		ctx := record.Context
		keys := make([]string, 0, len(ctx))
		for k := range ctx {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		if len(keys) > 0 {
			fmt.Fprint(buff, " [")
		}
		for i := 0; i < len(keys); i++ {
			k := keys[i]
			v := strconv.Quote(fmt.Sprintf("%+v", ctx[k]))
			if strings.IndexFunc(k, needsQuote) >= 0 || k == "" {
				k = strconv.Quote(k)
			}
			fmt.Fprintf(buff, "%s=%+v", k, v)
			if i < len(keys)-1 {
				buff.WriteString(" ")
			}
		}
		if len(keys) > 0 {
			fmt.Fprint(buff, "]")
		}
		fmt.Fprint(buff, "\n")
		return buff.Bytes()

	})
}

// Needs quote determines if provided rune is such that word that contains this
// rune needs to be quoted.
func needsQuote(r rune) bool {
	return r == ' ' || r == '"' || r == '\\' || r == '=' ||
		!unicode.IsPrint(r)
}

// JSONFormat is simple formatter that only marshals log record to json.
func JSONFormat(pretty bool) Formatter {
	return FormatterFunc(func(record Record) []byte {
		var marshaled []byte
		var err error
		if pretty {
			marshaled, err = json.MarshalIndent(record, "", "    ")
		} else {
			marshaled, err = json.Marshal(record)
		}
		if err != nil {
			marshaled, _ = json.Marshal(map[string]string{
				"JSONError": err.Error(),
			})
		}
		marshaled = append(marshaled, '\n')
		return marshaled
	})
}
