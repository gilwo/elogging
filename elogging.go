// package elogging provide enhanced logging capbilities with leveling and scope support
//
// # Levels
//
// levels are ordered : disabled (lowest - no output), error, warning, info, verbose, trace (highest)
//
// when setting a level, all lower levels are logged as well, higher levels are ignored from the log
//
// using Print(), Printf() or Println() is ignored from the leveled mechanism (they will be shown on the log output)
//
// # Log Objects
//
// all log objects are accessiable from the library and can me manipulated as well
//
// # Defaults
//
// when creating log objects, global defaults paramaters are set to each created log object.
// it is possible to change the log object paramters on the fly.
package elogging

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

const (
	ELSuppressRepeated = 1 << iota
	ELLikeDefaultLog
)

var (
	_stdLog       *Elog
	_defaultOut   io.Writer = os.Stderr
	_globalLevel  llevel
	logsActive    bool             = true
	_logs         map[*Elog]string = map[*Elog]string{}
	_defaultFlags int              = log.Ldate | log.Lmicroseconds | log.Llongfile | log.LUTC | log.Lmsgprefix /* Lshortfile override Llongfile */
	_elFlags      int              = ELLikeDefaultLog
)

func GetEloggingFlags() int {
	return _elFlags
}
func SetEloggingFlags(flags int) {
	_elFlags = flags
}
func checkElFlag(flags int) bool {
	return _elFlags&flags != 0
}

// DefaultFlags return the currently active flags for a new Elog
func DefaultFlags() int {
	return _defaultFlags
}

// SetDefaultFlags replace the default flags with the given flags value
func SetDefaultOutput(out io.Writer) {
	_defaultOut = out
}
func GetDefaultOutput() io.Writer {
	return _defaultOut
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

const (
	LEVEL_Disabled = "disabled"
	LEVEL_Error    = "error"
	LEVEL_Warning  = "warning"
	LEVEL_Info     = "info"
	LEVEL_Verbose  = "verbose"
	LEVEL_Trace    = "trace"
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

	case "err", "error":
		return lError
	case "wrn", "warn", "warning":
		return lWarn
	case "info", "inf":
		return lInfo
	case "verbose", "vrb":
		return lVerbose
	case "trace", "trc":
		return lTrace
	}
	return lDisabled
}

func _valid(level string) string {
	switch strings.ToLower(level) {
	case "err", "error":
		return "ERROR"
	case "wrn", "warn", "warning":
		return "WARN"
	case "info", "inf":
		return "INFO"
	case "verbose", "vrb":
		return "VERBOSE"
	case "trace", "trc":
		return "TRACE"
	}
	return "DISABLE"
}

// Elog represent a scoped leveled log
type Elog struct {
	scope       string
	level       llevel
	_log        *log.Logger
	_id         string
	_out        io.Writer
	__lastMsg   string
	__lastCount int
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

// SetGlobalLogLevel change the log level of all the Elog objects
func SetGlobalLogLevel(level string) {
	_globalLevel = _value(_valid(level))
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

type elogList []*Elog

func (a elogList) Len() int           { return len(a) }
func (a elogList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a elogList) Less(i, j int) bool { return a[i].scope < a[j].scope }

// ListScopedLogs return a list of all the existing Elog objects (sorted)
func ListScopedLogs() (elogs []*Elog) {
	for k := range _logs {
		elogs = append(elogs, k)
	}
	sort.Sort(elogList(elogs))
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

// GetScopedLogByID return the Elog object associated with the given ID
func GetScopedLogByID(id string) (elog *Elog) {
	for k := range _logs {
		if k._id == id {
			return k
		}
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
//
//	log.Ldate | log.Lmicroseconds | log.Llongfile | log.LUTC | log.Lmsgprefix
//
// level is the initial level for this log, empty level default to info level.
// out is where the log will be output, empty out default to os.stdout.
// check golang log packge doc for additional information.
func NewElog(scope, level string, out io.Writer) (e *Elog) {
	if out == nil {
		out = os.Stdout
	}
	if level == "" {
		level = "info"
	}
	// fmt.Printf("----- creating log with flags: [%#x]\n", _defaultFlags)
	e = &Elog{
		scope: scope,
		level: _value(_valid(level)),
		_log:  log.New(out, scope, _defaultFlags),
		_out:  out,
	}

	e._id = _hash(fmt.Sprintf("%s%p", scope, e))

	_logs[e] = scope
	return
}

// SetOutput allow to change the parameters of the log; output, level and output, previous log messages are not kept if output is changed
func (e *Elog) ModifyParams(modScope, modLevel string, modOut io.Writer) *Elog {
	if modScope != "" && modScope != e.scope {
		e.scope = modScope
		e._log.SetPrefix(modScope)
	}
	if modOut != nil && modOut != e._out {
		e._out = modOut
		e._log.SetOutput(modOut)
	}
	if modLevel != "" && modLevel != e.level.String() {
		e.level = _value(_valid(modLevel))
	}
	e._id = _hash(fmt.Sprintf("%s%p", e.scope, e))
	_logs[e] = e.scope
	return e
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

// CycleLevelUp change the current level of the Elog to the next level in a cyclic manner
func (e *Elog) CycleLevelUp() {
	e.level = (e.level + 1) % (lTrace + 1)
}

// CycleLevelDown change the current level of the Elog to the previous level in a cyclic manner
func (e *Elog) CycleLevelDown() {
	e.level = (e.level - 1) % (lTrace + 1)
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

func (e *Elog) Condf(condition bool, trueLevel, falseLevel, format string, args ...interface{}) {
	if condition {
		e._logf(_value(_valid(trueLevel)), format, args...)
	} else {
		e._logf(_value(_valid(falseLevel)), format, args...)
	}
}
func (e *Elog) Cond(condition bool, trueLevel, falseLevel string, args ...interface{}) {
	if condition {
		e._logf(_value(_valid(trueLevel)), "", args...)
	} else {
		e._logf(_value(_valid(falseLevel)), "", args...)
	}
}

func (e *Elog) _logf(level llevel, format string, args ...interface{}) {
	if !(logsActive && (level <= e.level || (_globalLevel > lDisabled && level <= _globalLevel))) {
		return
	}

	header := " (" + _valid(level.String()) + ") "
	msg := ""
	if format == "" {
		msg = header + fmt.Sprint(args...)
	} else {
		msg = header + fmt.Sprintf(format, args...)
	}

	if !checkElFlag(ELSuppressRepeated) {
		e._log.Output(3, msg)
		return
	}

	if msg == e.__lastMsg {
		e.__lastCount += 1
		msg = fmt.Sprintf(" last message repeated %d times", e.__lastCount)
	} else {
		e.__lastCount = 0
		e.__lastMsg = msg
	}

	if isPowerOfThree(e.__lastCount) || e.__lastCount == 0 {
		if e.__lastCount > 9 {
			msg += " (too many times)"
		}
		e._log.Output(3, msg)
	}
}

func isPowerOfThree(n int) bool {

	ansFloat := math.Log(float64(n)) / math.Log(3.0)
	ansInt := int(ansFloat)

	//Rounding to the 9th digit
	ansFloat = math.Round(ansFloat*1000000000) / 1000000000

	return ansFloat == float64(ansInt)
}

func init() {
	_stdLog = &Elog{
		scope: "",
		level: lTrace,
		_log:  log.Default(),
		_out:  os.Stderr,
	}

	_stdLog._id = _hash(fmt.Sprintf("%s%p", _stdLog.scope, _stdLog))

	_logs[_stdLog] = _stdLog.scope
}

// Println - same behavior as in original log when internal behaviour is propogate
func Println(args ...interface{}) {
	if checkElFlag(ELLikeDefaultLog) {
		_stdLog._log.Println(args...)
		return
	}
	if !logsActive || _stdLog.level == lDisabled {
		return
	}
	_stdLog._log.Output(2, " (Println) "+fmt.Sprintln(args...))
}

// Printf - same behavior as in original log when internal behaviour is propogate
func Printf(format string, args ...interface{}) {
	if checkElFlag(ELLikeDefaultLog) {
		_stdLog._log.Printf(format, args...)
		return
	}
	if !logsActive || _stdLog.level == lDisabled {
		return
	}
	_stdLog._log.Output(2, " (Printf) "+fmt.Sprintf(format, args...))
}

// Print - same behavior as in original log when internal behaviour is propogate
func Print(args ...interface{}) {
	if checkElFlag(ELLikeDefaultLog) {
		_stdLog._log.Print(args...)
		return
	}
	if !logsActive || _stdLog.level == lDisabled {
		return
	}
	_stdLog._log.Output(2, " (Print) "+fmt.Sprint(args...))
}

func Fatal(args ...interface{}) {
	_stdLog._log.Fatal(args...)
}
func Fatalf(format string, args ...interface{}) {
	_stdLog._log.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	_stdLog._log.Fatalln(args...)
}

func Panic(args ...interface{}) {
	_stdLog._log.Panic(args...)
}
func Panicf(format string, args ...interface{}) {
	_stdLog._log.Panicf(format, args...)
}
func Panicln(args ...interface{}) {
	_stdLog._log.Panicln(args...)
}

func DefaultLog() *Elog {
	return _stdLog
}

// util

func _hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
