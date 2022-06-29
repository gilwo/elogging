package elogging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	elog := NewElog("TestCreate", "info", os.Stdout)
	scopes, _, _ := ListScopesAndLevels()
	if !strings.Contains(strings.Join(scopes, " ; "), "TestCreate") {
		t.Error("list of logs does not have \"test\" as a scoped log")
	}
	elog.Info("a log line")
	elog.Clear()
}

func TestClear(t *testing.T) {
	defer func() {
		f := recover()
		if f != nil {
			if !strings.Contains(f.(error).Error(), "runtime error: invalid memory address or nil pointer dereference") {
				t.Errorf("unexpected error - %s\n", f.(error).Error())
			}
		}
	}()
	elog := NewElog("TestClear", "info", os.Stdout)
	elog.Error("problem just for the kicks")
	elog.Clear()
	elog.Print("a log line")
}

func TestLogToBuf(t *testing.T) {
	buf := []byte{}
	b := bytes.NewBuffer(buf)

	elog := NewElog("TestLogToBuf", "info", b)
	elog.Error("error message")
	elog.Warn("warning message")
	elog.Info("info message")
	elog.Verbose("verbose message")
	elog.SetLevel("trace")
	elog.Trace("trace message")
	bufMsg := b.String()
	if strings.Contains(bufMsg, "verbose message") {
		t.Error("unexpcted verbose level message")
	}
	if !strings.Contains(bufMsg, "trace message") {
		t.Error("expcted trace level message")
	}
}
