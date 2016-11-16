package ligno

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	isatty "github.com/mattn/go-isatty"
)

// Formatter is interface for converting log record to string representation.
type Formatter interface {
	Format(record Record) []byte
}

// FormatterFunc is function type that implements Formatter interface.
type FormatterFunc func(Record) []byte

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
		buff := buffPool.Get()
		defer buffPool.Put(buff)
		buff.WriteString(record.Time.Format(DefaultTimeFormat))
		buff.WriteRune(' ')
		buff.WriteString(record.Message)
		buff.WriteRune(' ')
		if record.File != "" && record.Line > 0 {
			buff.WriteRune('[')
			buff.WriteString(record.File)
			buff.WriteRune(':')
			buff.WriteString(strconv.Itoa(record.Line))
			buff.WriteRune(']')
		}
		buff.WriteRune('\n')
		return buff.Bytes()
	})
}

// TerminalFormat returns ThemeTerminalFormat with default theme set.
func TerminalFormat() Formatter {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return ThemedTerminalFormat(DefaultTheme)
	}
	return ThemedTerminalFormat(NoColorTheme)
}

// ThemedTerminalFormat returns formatter that produces records formatted for
// easy reading in terminal, but that are a bit richer then SimpleFormat (this
// one includes context keys)
func ThemedTerminalFormat(theme Theme) Formatter {
	return FormatterFunc(func(record Record) []byte {
		//time := record.Time.Format(DefaultTimeFormat)
		buff := buffPool.Get()
		defer buffPool.Put(buff)
		//buff.WriteString(theme.Time(time))
		//buff.WriteRune(' ')
		levelColor := theme.ForLevel(record.Level)
		levelName := record.Level.String()
		buff.WriteString(levelColor(levelName))
		padSpaces := levelNameMaxLength - len(levelName) + 2
		buff.Write(bytes.Repeat([]byte(" "), padSpaces))
		buff.WriteRune(' ')

		buff.WriteString(record.Message)

		ctx := record.Context
		keys := make([]string, 0, len(ctx))
		for k := range ctx {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		if len(keys) > 0 {
			buff.WriteString(" [")
		}
		for i := 0; i < len(keys); i++ {
			k := keys[i]
			keyQuote := strings.IndexFunc(k, needsQuote) >= 0 || k == ""
			if keyQuote {
				buff.WriteRune('"')
			}
			buff.WriteString(k)
			if keyQuote {
				buff.WriteRune('"')
			}
			buff.WriteRune('=')
			buff.WriteRune('"')
			buff.WriteString(fmt.Sprintf("%+v", ctx[k]))
			buff.WriteRune('"')
			if i < len(keys)-1 {
				buff.WriteRune(' ')
			}
		}
		if len(keys) > 0 {
			buff.WriteRune(']')
		}
		buff.WriteRune('\n')
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
		// since errors are not JSON serializable, make sure that all errors
		// are converted to strings
		for k, v := range record.Context {
			record.Context[k] = fmt.Sprintf("%+v", v)
		}

		// serialize
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
