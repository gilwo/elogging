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

func TestSuppress(t *testing.T) {

	buf := []byte{}
	b := bytes.NewBuffer(buf)
	SetEloggingFlags(GetEloggingFlags() | ELSuppressRepeated)
	elog := NewElog("TestLogSuppress", "info", b)
	for i := 0; i < 27; i++ {
		elog.Info("same message")
	}
	bufMsg := b.String()
	if strings.Count(bufMsg, "last message repeated") != 3 {
		t.Error("mismatch repeated count")
	}
	b.Reset()
	for i := 0; i < 82; i++ {
		elog.Info("same message2")
	}
	bufMsg = b.String()
	if strings.Count(bufMsg, "(too many times)") != 2 {
		t.Error("mismatch too many count")
	}
	b.Reset()
	SetEloggingFlags(GetEloggingFlags() & ^ELSuppressRepeated)
	for i := 0; i < 82; i++ {
		elog.Info("same message3")
	}
	bufMsg = b.String()
	if strings.Count(bufMsg, "same message3") != 82 {
		t.Error("mismatch repeated count")
	}
}

func TestStructured(t *testing.T) {
	buf := []byte{}
	b := bytes.NewBuffer(buf)
	SetEloggingFlags(GetEloggingFlags() | ELStructuredLog)
	elog := NewElog("TestLogStructured", "info", b)
	elog.Infof("message", "number", 1, "string", "st", "slice", []string{"1", "2"}, "map", map[string]int{"a": 1, "b": 2})
	bufMsg := b.String()
	t.Logf("output\n")
	t.Logf(bufMsg)
	b.Reset()

	elog.Infof("message2", "number")
	bufMsg = b.String()
	t.Logf("output2\n")
	t.Logf(bufMsg)
	b.Reset()

	elog.Infof("message3", 1, 2, 3)
	bufMsg = b.String()
	t.Logf("output3\n")
	t.Logf(bufMsg)
	b.Reset()

	elog.Infof("message4", 4)
	bufMsg = b.String()
	t.Logf("output4\n")
	t.Logf(bufMsg)
	b.Reset()
}
