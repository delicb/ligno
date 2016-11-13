package ligno

import (
	"fmt"
	"github.com/fatih/color"
	"strconv"
	"sync"
)

// Level represents rank of message importance.
// Log records can contain level and filters can decide not to process
// records based on this level.
type Level uint

// Levels of log records. Additional can be created, these are just defaults
// offered by library.
const (
	NOTSET   Level = iota
	DEBUG          = iota * 10
	INFO           = iota * 10
	WARNING        = iota * 10
	ERROR          = iota * 10
	CRITICAL       = iota * 10
)

var (
	// level2Name is map from level to name of known level names.
	level2Name = make(map[Level]string)
	// name2level is map from name to level of known level names.
	name2Level         = make(map[string]Level)
	levelNameMaxLength = 0
	mu                 sync.RWMutex
)

func init() {
	AddLevel("NOTSET", NOTSET)
	AddLevel("DEBUG", DEBUG)
	AddLevel("INFO", INFO)
	AddLevel("WARNING", WARNING)
	AddLevel("ERROR", ERROR)
	AddLevel("CRITICAL", CRITICAL)
}

// getLevelName returns name of provided level.
// If provided level does not exist, empty string is returned.
func getLevelName(level Level) (name string) {
	mu.RLock()
	defer mu.RUnlock()
	return level2Name[level]
}

// getLevelFromName returns level from provided name.
// If level with provided name is not found, NOTSET is returned.
func getLevelFromName(name string) (level Level) {
	mu.RLock()
	defer mu.RUnlock()
	return name2Level[name]
}

// AddLevel add new level to system with provided name and rank.
// It is forbidden to register levels that already exist.
func AddLevel(name string, rank Level) (Level, error) {
	mu.Lock()
	defer mu.Unlock()
	l := Level(rank)
	if _, ok := name2Level[name]; ok {
		return NOTSET, fmt.Errorf("level with name '%s' already exists", name)
	}
	if _, ok := level2Name[l]; ok {
		return NOTSET, fmt.Errorf("level with rank '%d' already exists", rank)
	}
	level2Name[l] = name
	name2Level[name] = l
	if len(name) > levelNameMaxLength {
		levelNameMaxLength = len(name)
	}
	return l, nil
}

// String returns level's string representation.
func (l Level) String() string {
	levelName := getLevelName(l)
	if levelName == "" {
		return fmt.Sprintf("Level(%d)", l)
	}
	return levelName
}

// MarshalJSON returns levels JSON representation (implementation of json.Marshaler)
func (l Level) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", l.String())), nil
}

// UnmarshalJSON recreates level from JSON representation (implementation of json.Unmarshaler)
func (l *Level) UnmarshalJSON(b []byte) error {
	levelStr, _ := strconv.Unquote(string(b))

	level, ok := name2Level[levelStr]
	if !ok {
		return fmt.Errorf("unknown level: %s", levelStr)
	}
	*l = level
	return nil
}

// Theme is definition of interface needed for colorizing log message to console.
type Theme interface {
	Time(msg string, args ...interface{}) string
	Debug(msg string, args ...interface{}) string
	Info(msg string, args ...interface{}) string
	Warning(msg string, args ...interface{}) string
	Error(msg string, args ...interface{}) string
	Critical(msg string, args ...interface{}) string
	ForLevel(level Level) func(msg string, args ...interface{}) string
}

type theme struct {
	timeColor     func(str string, args ...interface{}) string
	debugColor    func(str string, args ...interface{}) string
	infoColor     func(str string, args ...interface{}) string
	warningColor  func(str string, args ...interface{}) string
	errorColor    func(str string, args ...interface{}) string
	criticalColor func(str string, args ...interface{}) string
}

func (t *theme) Time(msg string, args ...interface{}) string {
	return t.timeColor(msg, args...)
}

func (t *theme) Debug(msg string, args ...interface{}) string {
	return t.debugColor(msg, args...)
}

func (t *theme) Info(msg string, args ...interface{}) string {
	return t.infoColor(msg, args...)
}

func (t *theme) Warning(msg string, args ...interface{}) string {
	return t.warningColor(msg, args...)
}

func (t *theme) Error(msg string, args ...interface{}) string {
	return t.errorColor(msg, args...)
}

func (t *theme) Critical(msg string, args ...interface{}) string {
	return t.criticalColor(msg, args...)
}

func (t *theme) ForLevel(level Level) func(msg string, args ...interface{}) string {
	switch {
	case level < INFO:
		return t.debugColor
	case level >= INFO && level < WARNING:
		return t.infoColor
	case level >= WARNING && level < ERROR:
		return t.warningColor
	case level >= ERROR && level < CRITICAL:
		return t.errorColor
	case level >= CRITICAL:
		return t.criticalColor
	default:
		return t.infoColor
	}
}

var (
	// DefaultTheme defines theme used by default.
	DefaultTheme = &theme{
		timeColor:     color.New(color.FgWhite, color.Faint).SprintfFunc(),
		debugColor:    color.New(color.FgWhite).SprintfFunc(),
		infoColor:     color.New(color.FgHiWhite).SprintfFunc(),
		warningColor:  color.New(color.FgYellow).SprintfFunc(),
		errorColor:    color.New(color.FgHiRed).SprintfFunc(),
		criticalColor: color.New(color.BgRed, color.FgHiWhite).SprintfFunc(),
	}

	// NoColorTheme defines theme that does not color any output.
	NoColorTheme = &theme{
		timeColor:     fmt.Sprintf,
		debugColor:    fmt.Sprintf,
		infoColor:     fmt.Sprintf,
		warningColor:  fmt.Sprintf,
		errorColor:    fmt.Sprintf,
		criticalColor: fmt.Sprintf,
	}
)
