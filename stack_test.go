package errors

import (
	stderrors "errors"
	"fmt"
	"github.com/confetti-framework/syslog/log_level"
	"github.com/stretchr/testify/assert"
	net "net/http"
	"runtime"
	"testing"
)

var initpc = caller()

type X struct{}

// val returns a Frame pointing to itself.
func (x X) val() Frame {
	return caller()
}

// ptr returns a Frame pointing to itself.
func (x *X) ptr() Frame {
	return caller()
}

func TestFrameFormat(t *testing.T) {
	var tests = []struct {
		Frame
		format string
		want   string
	}{{
		initpc,
		"%s",
		"stack_test.go",
	}, {
		initpc,
		"%+s",
		"github.com/confetti-framework/errors.init\n" +
			"\t.+/github.com/confetti-framework/errors/stack_test.go",
	}, {
		0,
		"%s",
		"unknown",
	}, {
		0,
		"%+s",
		"unknown",
	}, {
		initpc,
		"%d",
		"13",
	}, {
		0,
		"%d",
		"0",
	}, {
		initpc,
		"%n",
		"init",
	}, {
		func() Frame {
			var x X
			return x.ptr()
		}(),
		"%n",
		`\(\*X\).ptr`,
	}, {
		func() Frame {
			var x X
			return x.val()
		}(),
		"%n",
		"X.val",
	}, {
		0,
		"%n",
		"",
	}, {
		initpc,
		"%v",
		"stack_test.go:13",
	}, {
		initpc,
		"%+v",
		"github.com/confetti-framework/errors.init\n" +
			"\t.+/github.com/confetti-framework/errors/stack_test.go:13",
	}, {
		0,
		"%v",
		"unknown:0",
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.Frame, tt.format, tt.want)
	}
}

func TestFuncname(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"", ""},
		{"runtime.main", "main"},
		{"github.com/confetti-framework/errors.funcname", "funcname"},
		{"funcname", "funcname"},
		{"io.copyBuffer", "copyBuffer"},
		{"main.(*R).Write", "(*R).Write"},
	}

	for _, tt := range tests {
		got := funcname(tt.name)
		want := tt.want
		if got != want {
			t.Errorf("funcname(%q): want: %q, got %q", tt.name, want, got)
		}
	}
}

func TestStackTrace(t *testing.T) {
	tests := []struct {
		err  error
		want []string
	}{{
		New("ooh"), []string{
			"github.com/confetti-framework/errors.TestStackTrace\n" +
				"\t.+/github.com/confetti-framework/errors/stack_test.go:125",
		},
	}, {
		Wrap(New("ooh"), "ahh"), []string{
			"github.com/confetti-framework/errors.TestStackTrace\n" +
				"\t.+/github.com/confetti-framework/errors/stack_test.go:130", // this is the stack of Wrap, not New
		},
	}, {
		Unwrap(Wrap(New("ooh"), "ahh")), []string{
			"github.com/confetti-framework/errors.TestStackTrace\n" +
				"\t.+/github.com/confetti-framework/errors/stack_test.go:135", // this is the stack of New
		},
	}, {
		func() error { return New("ooh") }(), []string{
			`github.com/confetti-framework/errors.TestStackTrace.func1` +
				"\n\t.+/github.com/confetti-framework/errors/stack_test.go:140", // this is the stack of New
			"github.com/confetti-framework/errors.TestStackTrace\n" +
				"\t.+/github.com/confetti-framework/errors/stack_test.go:140", // this is the stack of New's caller
		},
	}, {
		Unwrap(func() error {
			return func() error {
				return New("hello %s", fmt.Sprintf("world: %s", "ooh"))
			}()
		}()), []string{
			`github.com/confetti-framework/errors.TestStackTrace.func2.1` +
				"\n\t.+/github.com/confetti-framework/errors/stack_test.go:149", // this is the stack of New
			`github.com/confetti-framework/errors.TestStackTrace.func2` +
				"\n\t.+/github.com/confetti-framework/errors/stack_test.go:150", // this is the stack of New's caller
			"github.com/confetti-framework/errors.TestStackTrace\n" +
				"\t.+/github.com/confetti-framework/errors/stack_test.go:151", // this is the stack of New's caller's caller
		},
	}}
	for i, tt := range tests {
		x, ok := tt.err.(interface {
			StackTrace() StackTrace
		})
		if !ok {
			t.Errorf("expected %#v to implement StackTrace() StackTrace", tt.err)
			continue
		}
		st := x.StackTrace()
		for j, want := range tt.want {
			testFormatRegexp(t, i, st[j], "%+v", want)
		}
	}
}

func stackTrace() StackTrace {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return st.StackTrace()
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		StackTrace
		format string
		want   string
	}{{
		nil,
		"%s",
		`\[\]`,
	}, {
		nil,
		"%v",
		`\[\]`,
	}, {
		nil,
		"%+v",
		"",
	}, {
		nil,
		"%#v",
		`\[\]errors.Frame\(nil\)`,
	}, {
		make(StackTrace, 0),
		"%s",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%v",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%+v",
		"",
	}, {
		make(StackTrace, 0),
		"%#v",
		`\[\]errors.Frame{}`,
	}, {
		stackTrace()[:2],
		"%s",
		`\[stack_test.go stack_test.go\]`,
	}, {
		stackTrace()[:2],
		"%v",
		`\[stack_test.go:178 stack_test.go:225\]`,
	}, {
		stackTrace()[:2],
		"%+v",
		"\n" +
			"github.com/confetti-framework/errors.stackTrace\n" +
			"\t.+/github.com/confetti-framework/errors/stack_test.go:178\n" +
			"github.com/confetti-framework/errors.TestStackTraceFormat\n" +
			"\t.+/github.com/confetti-framework/errors/stack_test.go:229",
	}, {
		stackTrace()[:2],
		"%#v",
		`\[\]errors.Frame{stack_test.go:178, stack_test.go:237}`,
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

func TestGetStackTraceFromStatusErr(t *testing.T) {
	err := New("message").Status(net.StatusNotFound)
	result := fmt.Sprintf("%+v", err)
	assert.Contains(t, result, "stack_test.go")
}

func TestGetStackTraceFromLevelErr(t *testing.T) {
	err := New("message").Level(log_level.ALERT)
	result := fmt.Sprintf("%+v", err)
	assert.Contains(t, result, "stack_test.go")
}

func TestGetStackTraceFromSimpleErr(t *testing.T) {
	err := WithStatus(stderrors.New("message"), net.StatusNotFound)
	result := fmt.Sprintf("%+v", err)
	assert.Equal(t, "message", result)
}

// a version of runtime.Caller that returns a Frame, not a uintptr.
func caller() Frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	return Frame(frame.PC)
}
