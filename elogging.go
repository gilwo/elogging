// enhanced logging library with leveling and scope support
//
// levels are ordered : disabled (lowest - no output), error, warning, info, verbose, trace (highest)
//
// when setting a level, all lower levels are logged as well, higher levels are ignored from the log
//
// using Print(), Printf() or Println() is ignored from the leveled mechanism (they will be shown on the log output)
package elogging

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	_defaultOut   io.Writer
	logsActive    bool             = true
	_logs         map[*Elog]string = map[*Elog]string{}
	_defaultFlags                  = log.Ldate | log.Lmicroseconds | log.Llongfile | log.LUTC | log.Lmsgprefix /* Lshortfile override Llongfile */
)

// DefaultFlags return the currently active flags for a new Elog
func DefaultFlags() int {
	return _defaultFlags
}

// SetDefaultFlags replace the default flags with the given flags value
func SetDefaultOutput(out io.Writer) {
	_defaultOut = out
}

// SetDefaultFlags replace the default flags with the given flags value
func SetDefaultFlags(flags int) {
	_defaultFlags = flags
}

// LogsOff disable all output logs from logs created by the logging library
func LogsOff() {
	logsActive = false
}

// LogsOn enable logs output, all levels are resumed to their previous levels
func LogsOn() {
	logsActive = true
}

type llevel int32

const (
	lDisabled llevel = iota
	lError
	lWarn
	lInfo
	lVerbose
	lTrace
)

func (l llevel) String() string {
	switch l {
	case lDisabled:
		return "Disabled"
	case lError:
		return "Error"
	case lWarn:
		return "Warning"
	case lInfo:
		return "Info"
	case lVerbose:
		return "Verbose"
	case lTrace:
		return "Trace"
	}
	return "Disabled"
}

func _value(level string) llevel {
	switch strings.ToLower(_valid(level)) {

	case "error":
		return lError
	case "warn":
		fallthrough
	case "warning":
		return lWarn
	case "info":
		return lInfo
	case "verbose":
		return lVerbose
	case "trace":
		return lTrace
	}
	return lDisabled
}

func _valid(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return "ERORR"
	case "warn":
		fallthrough
	case "warning":
		return "WARN"
	case "info":
		return "INFO"
	case "verbose":
		return "VERBOSE"
	case "trace":
		return "TRACE"
	}
	return "DISABLE"
}

// Elog represent a scoped leveled log
type Elog struct {
	scope string
	level llevel
	_log  *log.Logger
	_id   string
}

// String descrption of an Elog instance
func (e Elog) String() string {
	return fmt.Sprintf("[%s:%s:(%s)]", e._id, e.scope, e.level)
}

// Scope retrieve the scope of the given Elog instance
func (e *Elog) Scope() string {
	return e.scope
}

// ID retrieve the id of the given Elog instance
func (e *Elog) ID() string {
	return e._id
}

// SetScopeLogLevelByName change log level of a scope log to level, the scope logs is looked for its name.
// if multiple logs with the same scope name exits only the first occureance of scoped log will be changed
//
// TODO: not sure about this one
func SetScopeLogLevelByName(scope, level string) {
	for k, v := range _logs {
		if v == scope {
			k.SetLevel(level)
			return
		}
	}
}

// SetScopeLogLevelByID change the log level of the Elog associated with the given id
func SetScopeLogLevelByID(id, level string) {
	for k := range _logs {
		if k._id == id {
			k.SetLevel(level)
			return
		}
	}
}

// ListScopedLogs return a list of all the existing Elog
func ListScopedLogs() (elogs []*Elog) {
	for k := range _logs {
		elogs = append(elogs, k)
	}
	return
}

// ListScopesAndLevels return a lists of scopes, ids and levels for the existing logs
func ListScopesAndLevels() (scopes, ids, levels []string) {
	for k, v := range _logs {
		scopes = append(scopes, v)
		levels = append(levels, k.GetLevel())
		ids = append(ids, k._id)
	}
	return
}

// Create an Elog object
func NewElogDefaults(scope string) *Elog {
	return NewElog(scope, "info", _defaultOut)
}

// NewElog create a scoped leveled logger wrapping the native golang log package.
// it creates a new log and provide scheme to have a scope and level for the logger.
// the newly created logger is created with the following flags:
//  log.Ldate | log.Lmicroseconds | log.Llongfile | log.LUTC | log.Lmsgprefix
// level is the initial level for this log, empty level default to info level.
// out is where the log will be output, empty out default to os.stdout.
// check golang log packge doc for additional information.
func NewElog(scope, level string, out io.Writer) *Elog {
	if out == nil {
		out = os.Stdout
	}
	if level == "" {
		level = "info"
	}
	// fmt.Printf("----- creating log with flags: [%#x]\n", _defaultFlags)
	ret := &Elog{
		scope: scope,
		level: _value(_valid(level)),
		_log:  log.New(out, scope, _defaultFlags),
	}
	_hash := func(s string) string {
		h := sha1.New()
		h.Write([]byte(s))
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	ret._id = _hash(fmt.Sprintf("%s%p", scope, ret))

	_logs[ret] = scope
	return ret
}

// Clear remove this Elog from the existing Elog, the Elog is unsuable following this invocation
//
// log is invalid following this invocation and any additional calls will create an unexpected behaviour
//
// TODO: check what happens if someone call any leveled log for this log
func (e *Elog) Clear() {
	delete(_logs, e)
	e.level = lDisabled
	e._log = nil
}

// SetLevel change the current level of the Elog to the given level
func (e *Elog) SetLevel(level string) {
	e.level = _value(_valid(level))
}

// GetLevel retrieve the current level of the Elog
func (e *Elog) GetLevel() string {
	return e.level.String()
}

// GetFlags retrieve the current flags of the Elog
func (e *Elog) GetFlags() int {
	return e._log.Flags()
}

// SetFlags replace the current flags of the Elog
func (e *Elog) SetFlags(flags int) {
	e._log.SetFlags(flags)
}

// Println print prefixed (Println) log lines ingoring the leveled logging mechanism
func (e *Elog) Println(args ...interface{}) {
	if !logsActive {
		return
	}
	e._log.Output(2, " (Println) "+fmt.Sprintln(args...))
}

// Printf print prefixed (Printf) log lines ingoring the leveled logging mechanism
func (e *Elog) Printf(format string, args ...interface{}) {
	if !logsActive {
		return
	}
	e._log.Output(2, " (Printf) "+fmt.Sprintf(format, args...))
}

// Print print prefixed (Print) log lines ingoring the leveled logging mechanism
func (e *Elog) Print(args ...interface{}) {
	if !logsActive {
		return
	}
	e._log.Output(2, " (Print) "+fmt.Sprint(args...))
}

// All methods below are relate to the level logging mechanism

// Errorf print prefixed (Error) formatted log lines with level Error
func (e *Elog) Errorf(format string, args ...interface{}) {
	e._logf(lError, format, args...)
}

// Warnf print prefixed (Warning) formatted log lines with level Warning
func (e *Elog) Warnf(format string, args ...interface{}) {
	e._logf(lWarn, format, args...)
}

// Infof print prefixed (Info) formatted log lines with level Info
func (e *Elog) Infof(format string, args ...interface{}) {
	e._logf(lInfo, format, args...)
}

// Verbosef print prefixed (Verbose) formatted log lines with level Verbose
func (e *Elog) Verbosef(format string, args ...interface{}) {
	e._logf(lVerbose, format, args...)
}

// Tracef print prefixed (Trace) formatted log lines with level Trace
func (e *Elog) Tracef(format string, args ...interface{}) {
	e._logf(lTrace, format, args...)
}

// Errorf print prefixed (Error) log lines with level Error
func (e *Elog) Error(args ...interface{}) {
	e._logf(lError, "", args...)
}

// Warn print prefixed (Warning) log lines with level Warning
func (e *Elog) Warn(args ...interface{}) {
	e._logf(lWarn, "", args...)
}

// Info print prefixed (Info) log lines with level Info
func (e *Elog) Info(args ...interface{}) {
	e._logf(lInfo, "", args...)
}

// Verbose print prefixed (Verbose) log lines with level Verbose
func (e *Elog) Verbose(args ...interface{}) {
	e._logf(lVerbose, "", args...)
}

// Trace print prefixed (Trace) log lines with level Trace
func (e *Elog) Trace(args ...interface{}) {
	e._logf(lTrace, "", args...)
}

func (e *Elog) _logf(level llevel, format string, args ...interface{}) {
	if level > e.level || !logsActive {
		return
	}
	header := " (" + _valid(level.String()) + ") "
	if format == "" {
		e._log.Output(3, header+fmt.Sprint(args...))
	} else {
		e._log.Output(3, header+fmt.Sprintf(format, args...))
	}
}
