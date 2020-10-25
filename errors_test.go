package errors

import (
	stderrors "errors"
	"fmt"
	"github.com/lanvard/syslog/log_level"
	"github.com/stretchr/testify/assert"
	"io"
	net "net/http"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		err  string
		want error
	}{
		{"", fmt.Errorf("")},
		{"foo", fmt.Errorf("foo")},
		{"foo", New("foo")},
		{"string with format specifiers: %v", stderrors.New("string with format specifiers: %v")},
	}

	for _, tt := range tests {
		got := New(tt.err)
		if got.Error() != tt.want.Error() {
			t.Errorf("New.Error(): got: %q, want %q", got, tt.want)
		}
	}
}

func TestFundamentalNewWithString(t *testing.T) {
	err := New("not found")
	assert.Equal(t, "not found", err.Error())
}

func TestFundamentalNewWithArgument(t *testing.T) {
	err := New("%s not found", "user")
	assert.Equal(t, "user not found", err.Error())
}

func TestFundamentalNewWithArguments(t *testing.T) {
	err := New("%s not found in %s", "user", "account")
	assert.Equal(t, "user not found in account", err.Error())
}

func TestFundamentalFluentWrap(t *testing.T) {
	assert.Equal(t, "database error: not found", New("not found").Wrap("database error").Error())
}

func TestFundamentalFluentLevel(t *testing.T) {
	err := New("database error").Level(log_level.DEBUG)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
}

func TestFundamentalFluentWrapFluentLevel(t *testing.T) {
	wrapper := New("database error").Wrap("system error")
	err := wrapper.Level(log_level.DEBUG)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
	assert.Equal(t, "system error: database error", err.Error())
}

func TestFundamentalFluentWrapFluentWrap(t *testing.T) {
	wrapper := New("database error").Wrap("system error").Wrap("attention")
	err := wrapper.Level(log_level.DEBUG)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
	assert.Equal(t, "attention: system error: database error", err.Error())
}

func TestFundamentalFluentStatus(t *testing.T) {
	err := New("not found").Status(net.StatusBadRequest)
	assert.Equal(t, "database error: not found", err.Wrap("database error").Error())

	level, ok := FindStatus(err.Wrap("database error"))
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

func TestWrapNil(t *testing.T) {
	got := Wrap(nil, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrapFormat(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Wrap(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrap(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestWrapFluentLevel(t *testing.T) {
	wrapper := Wrap(New("database error"), "system error")
	err := wrapper.Level(log_level.ERROR)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.ERROR, level)
	assert.Equal(t, "system error: database error", err.Error())
}

func TestWrapFluentStatus(t *testing.T) {
	err := Wrap(New("not found"), "database error").Status(net.StatusBadRequest)
	assert.Equal(t, "database error: not found", err.Error())

	level, ok := FindStatus(err)
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

func TestLevelWithoutLevel(t *testing.T) {
	err := New("database error")

	level, ok := FindLevel(err)
	assert.False(t, ok)
	assert.Equal(t, log_level.EMERGENCY, level)
}

func TestLevelWithNil(t *testing.T) {
	assert.Nil(t, WithLevel(nil, log_level.DEBUG))
}

func TestLevelWithEmergency(t *testing.T) {
	err := WithLevel(New("database error"), log_level.EMERGENCY)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.EMERGENCY, level)
}

func TestLevelWithDebug(t *testing.T) {
	err := WithLevel(New("database error"), log_level.DEBUG)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
}

func TestLevelFromCause(t *testing.T) {
	var err error
	err = WithLevel(New("database error"), log_level.DEBUG)
	err = Wrap(err, "system error")

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
}

func TestLevelFluentWrap(t *testing.T) {
	err := WithLevel(New("not found"), log_level.DEBUG)
	assert.Equal(t, "database error: not found", err.Wrap("database error").Error())

	level, ok := FindLevel(err.Wrap("database error"))
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
}

func TestLevelFluentLevel(t *testing.T) {
	wrapper := WithLevel(New("database error"), log_level.DEBUG)
	err := wrapper.Level(log_level.DEBUG)

	level, ok := FindLevel(err)
	assert.True(t, ok)
	assert.Equal(t, log_level.DEBUG, level)
	assert.Equal(t, "database error", err.Error())
}

func TestLevelFluentStatus(t *testing.T) {
	err := WithLevel(New("not found"), log_level.EMERGENCY).Status(net.StatusBadRequest)
	assert.Equal(t, "database error: not found", err.Wrap("database error").Error())

	level, ok := FindStatus(err.Wrap("database error"))
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

func TestStatusWithoutStatus(t *testing.T) {
	err := New("database error")

	level, ok := FindStatus(err)
	assert.False(t, ok)
	assert.Equal(t, net.StatusInternalServerError, level)
}

func TestStatusWithNil(t *testing.T) {
	assert.Nil(t, WithStatus(nil, net.StatusInternalServerError))
}

func TestStatusWithEmergency(t *testing.T) {
	err := WithStatus(New("database error"), net.StatusInternalServerError)

	level, ok := FindStatus(err)
	assert.True(t, ok)
	assert.Equal(t, net.StatusInternalServerError, level)
}

func TestStatusWithDebug(t *testing.T) {
	err := WithStatus(New("database error"), net.StatusBadRequest)

	level, ok := FindStatus(err)
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

func TestStatusFromCause(t *testing.T) {
	var err error
	err = WithStatus(New("database error"), net.StatusBadRequest)
	err = Wrap(err, "system error")

	level, ok := FindStatus(err)
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

func TestStatusFluentWrap(t *testing.T) {
	err := WithStatus(New("not found"), net.StatusBadRequest)
	assert.Equal(t, "database error: not found", err.Wrap("database error").Error())

	level, ok := FindStatus(err.Wrap("database error"))
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

func TestStatusFluentStatus(t *testing.T) {
	err := WithStatus(New("not found"), net.StatusInternalServerError).Status(net.StatusBadRequest)
	assert.Equal(t, "database error: not found", err.Wrap("database error").Error())

	level, ok := FindStatus(err.Wrap("database error"))
	assert.True(t, ok)
	assert.Equal(t, net.StatusBadRequest, level)
}

type nilError struct{}

func (nilError) Error() string { return "nil error" }

func TestCause(t *testing.T) {
	x := New("error")
	tests := []struct {
		err  error
		want error
	}{{
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		// explicit nil error is nil
		err:  (error)(nil),
		want: nil,
	}, {
		// typed nil is nil
		err:  (*nilError)(nil),
		want: (*nilError)(nil),
	}, {
		// uncaused error is unaffected
		err:  io.EOF,
		want: io.EOF,
	}, {
		// caused error returns cause
		err:  Wrap(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		err:  x, // return from errors.New
		want: x,
	}, {
		WithMessage(nil, "whoops"),
		nil,
	}, {
		WithMessage(io.EOF, "whoops"),
		io.EOF,
	}, {
		WithStack(nil),
		nil,
	}, {
		WithStack(io.EOF),
		io.EOF,
	}}

	for i, tt := range tests {
		got := Cause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func TestWrapFormatNil(t *testing.T) {
	got := Wrap(nil, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Wrap(io.EOF, "read error without format specifiers"), "client error", "client error: read error without format specifiers: EOF"},
		{Wrap(io.EOF, "read error with %d format specifier", 1), "client error", "client error: read error with 1 format specifier: EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrap(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{New("read error without format specifiers"), "read error without format specifiers"},
		{New("read error with %d format specifier", 1), "read error with 1 format specifier"},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("New()(%v): got: %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestWithStackNil(t *testing.T) {
	got := WithStack(nil)
	if got != nil {
		t.Errorf("WithStack(nil): got %#v, expected nil", got)
	}
}

func TestWithStack(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{WithStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := WithStack(tt.err).Error()
		if got != tt.want {
			t.Errorf("WithStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestWithMessageNil(t *testing.T) {
	got := WithMessage(nil, "no error").Cause()
	if got != nil {
		t.Errorf("WithMessage(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWithMessage(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessage(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := WithMessage(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestWithMessagefNil(t *testing.T) {
	got := WithMessage(nil, "no error")
	if got.Cause() != nil {
		t.Errorf("WithMessage(nil, \"no error\"): got %#v, expected nil", got.Cause())
	}
	assert.Equal(t, "no error", got.Error())
}

func TestWithMessagef(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessage(io.EOF, "read error without format specifier"), "client error", "client error: read error without format specifier: EOF"},
		{WithMessage(io.EOF, "read error with %d format specifier", 1), "client error", "client error: read error with 1 format specifier: EOF"},
	}

	for _, tt := range tests {
		got := WithMessage(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

// errors.New, etc values are not expected to be compared by value
// but the change in errors#27 made them incomparable. Assert that
// various kinds of errors have a functional equality operator, even
// if the result of that equality is always false.
func TestErrorEquality(t *testing.T) {
	vals := []error{
		nil,
		io.EOF,
		stderrors.New("EOF"),
		New("EOF"),
		New("EOF"),
		Wrap(io.EOF, "EOF"),
		Wrap(io.EOF, "EOF%d", 2),
		WithMessage(nil, "whoops"),
		WithMessage(io.EOF, "whoops"),
		WithStack(io.EOF),
		WithStack(nil),
	}

	for i := range vals {
		for j := range vals {
			_ = vals[i] == vals[j] // mustn't panic
		}
	}
}
