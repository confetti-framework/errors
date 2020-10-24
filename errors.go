// Package errors provides simple error handling primitives.
//
// The traditional error handling idiom in Go is roughly akin to
//
//     if err != nil {
//             return err
//     }
//
// which when applied recursively up the call stack results in error reports
// without context or debugging information. The errors package allows
// programmers to add context to the failure path in their code in a way
// that does not destroy the original value of the error.
//
// Adding context to an error
//
// The errors.Wrap function returns a new error that adds context to the
// original error by recording a stack trace at the point Wrap is called,
// together with the supplied message. For example
//
//     _, err := ioutil.ReadAll(r)
//     if err != nil {
//             return errors.Wrap(err, "read failed")
//     }
//
// If additional control is required, the errors.WithStack and
// errors.WithMessage functions destructure errors.Wrap into its component
// operations: annotating an error with a stack trace and with a message,
// respectively.
//
// Retrieving the cause of an error
//
// Using errors.Wrap constructs a stack of errors, adding context to the
// preceding error. Depending on the nature of the error it may be necessary
// to reverse the operation of errors.Wrap to retrieve the original error
// for inspection. Any error value which implements this interface
//
//     type causer interface {
//             Cause() error
//     }
//
// can be inspected by errors.Cause. errors.Cause will recursively retrieve
// the topmost error that does not implement causer, which is assumed to be
// the original cause. For example:
//
//     switch err := errors.Cause(err).(type) {
//     case *MyError:
//             // handle specifically
//     default:
//             // unknown error
//     }
//
// Although the causer interface is not exported by this package, it is
// considered a part of its stable public interface.
//
// Formatted printing of errors
//
// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported:
//
//     %s    print the error. If the error has a Cause it will be
//           printed recursively.
//     %v    see %s
//     %+v   extended format. Each Frame of the error's StackTrace will
//           be printed in detail.
//
// Retrieving the stack trace of an error or wrapper
//
// New, errorf, Wrap, and Wrap record a stack trace at the point they are
// invoked. This information can be retrieved with the following interface:
//
//     type stackTracer interface {
//             StackTrace() errors.StackTrace
//     }
//
// The returned errors.StackTrace type is defined as
//
//     type StackTrace []Frame
//
// The Frame type represents a call site in the stack trace. Frame supports
// the fmt.Formatter interface that can be used for printing information about
// the stack trace of this error. For example:
//
//     if err, ok := err.(stackTracer); ok {
//             for _, f := range err.StackTrace() {
//                     fmt.Printf("%+s:%d\n", f, f)
//             }
//     }
//
// Although the stackTracer interface is not exported by this package, it is
// considered a part of its stable public interface.
//
// See the documentation for Frame.Format for more details.
package errors

import (
	"fmt"
	syslog "github.com/lanvard/syslog/level"
	"io"
)

// New returns an error with the supplied message and formats
// according to a format specifier and returns the string
// as a value that satisfies error.
// New also records the stack trace at the point it was called.
func New(message string, args ...interface{}) *fundamental {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return &fundamental{
		msg:   message,
		stack: callers(),
	}
}

// fundamental is an error that has a message and a stack, but no caller.
type fundamental struct {
	msg string
	*stack
}

func (f *fundamental) Error() string { return f.msg }

func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, f.msg)
			f.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, f.msg)
	case 'q':
		fmt.Fprintf(s, "%q", f.msg)
	}
}

func WithLevel(err error, level syslog.Level) *withLevel {
	if err == nil {
		return nil
	}
	return &withLevel{
		err,
		level,
	}
}

type withLevel struct {
	cause error
	level syslog.Level
}

func (w *withLevel) Error() string {
	return w.cause.Error()
}

func (w *withLevel) Cause() error { return w.cause }

func (w *withLevel) Wrap(message string, args ...interface{}) error {
	return WithMessage(w, message, args...)
}

func (w *withLevel) Level(level syslog.Level) *withLevel {
	return WithLevel(w, level)
}

func FindLevel(err error) syslog.Level {
	var level syslog.Level

	if level, ok := err.(*withLevel); ok {
		return level.level
	}

	if causer, ok := err.(causer); ok {
		return FindLevel(causer.Cause())
	}
	return level
}

func (f *fundamental) Wrap(message string, args ...interface{}) *withMessage {
	return WithMessage(f, message, args...)
}

func (f *fundamental) Level(level syslog.Level) *withLevel {
	return WithLevel(f, level)
}

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &withStack{
		err,
		callers(),
	}
}

type withStack struct {
	error
	*stack
}

func (w *withStack) Cause() error { return w.error }

func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", w.Cause())
			w.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

func (w *withStack) Level(level syslog.Level) *withLevel {
	return WithLevel(w, level)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string, args ...interface{}) *withStack {
	if err == nil {
		return nil
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	err = &withMessage{
		cause: err,
		msg:   message,
	}
	return &withStack{
		err,
		callers(),
	}
}

// WithMessage annotates err with a new message.
func WithMessage(err error, message string, args ...interface{}) *withMessage {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}

type withMessage struct {
	cause error
	msg   string
}

func (w *withMessage) Error() string {
	if w.cause == nil {
		return w.msg
	}
	return w.msg + ": " + w.cause.Error()
}

func (w *withMessage) Cause() error { return w.cause }

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

func (w *withMessage) Level(level syslog.Level) *withLevel {
	return WithLevel(w, level)
}

func (w *withMessage) Wrap(message string, args ...interface{}) *withMessage {
	return WithMessage(w, message, args...)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

type causer interface {
	Cause() error
}
