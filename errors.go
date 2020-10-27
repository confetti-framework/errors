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
//     type unwrapper interface {
//             Unwrap() error
//     }
//
// can be inspected by errors.Unwrap. errors.Unwrap will recursively retrieve
// the topmost error that does not implement unwrapper, which is assumed to be
// the original cause. For example:
//
//     switch err := errors.Unwrap(err).(type) {
//     case *MyError:
//             // handle specifically
//     default:
//             // unknown error
//     }
//
// Although the unwrapper interface is not exported by this package, it is
// considered a part of its stable public interface.
//
// Formatted printing of errors
//
// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported:
//
//     %s    print the error. If the error has a Unwrap it will be
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
	syslog "github.com/lanvard/syslog/log_level"
	"io"
	net "net/http"
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

func (f *fundamental) Error() string {
	return f.msg
}

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

func (f *fundamental) StackTrace() StackTrace {
	return f.stack.StackTrace()
}

func (f *fundamental) Wrap(message string, args ...interface{}) *withMessage {
	return WithMessage(f, message, args...)
}

func (f *fundamental) Level(level syslog.Level) *withLevel {
	return WithLevel(f, level)
}

func (f *fundamental) Status(status int) *withStatus {
	return WithStatus(f, status)
}

func FindLevel(err error) (syslog.Level, bool) {
	var level syslog.Level
	var levelHolder *withLevel

	if !As(err, &levelHolder) {
		return level, false
	}

	return levelHolder.level, true
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

func (w *withLevel) Unwrap() error {
	return w.cause
}

func (w *withLevel) Wrap(message string, args ...interface{}) error {
	return WithMessage(w, message, args...)
}

func (w *withLevel) Level(level syslog.Level) *withLevel {
	return WithLevel(w, level)
}

func (w *withLevel) Status(status int) *withStatus {
	return WithStatus(w, status)
}

func FindStatus(err error) (int, bool) {
	var statusHolder *withStatus

	ok := As(err, &statusHolder)
	if !ok {
		return net.StatusInternalServerError, false
	}

	return statusHolder.status, true
}

func WithStatus(err error, status int) *withStatus {
	if err == nil {
		return nil
	}
	return &withStatus{
		err,
		status,
	}
}

type withStatus struct {
	cause  error
	status int
}

func (w *withStatus) Error() string {
	return w.cause.Error()
}

func (w *withStatus) Unwrap() error { return w.cause }

func (w *withStatus) Wrap(message string, args ...interface{}) error {
	return WithMessage(w, message, args...)
}

func (w *withStatus) Level(status syslog.Level) *withLevel {
	return WithLevel(w, status)
}

func (w *withStatus) Status(status int) *withStatus {
	return WithStatus(w, status)
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

func FindStack(err error) (StackTrace, bool) {
	var stackHolder interface{ StackTrace() StackTrace }

	if !As(err, &stackHolder) {
		return StackTrace{}, false
	}

	return stackHolder.StackTrace(), true
}

type withStack struct {
	error
	*stack
}

func (w *withStack) Unwrap() error { return w.error }

func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", w.Unwrap())
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

func (w *withStack) StackTrace() StackTrace {
	return w.stack.StackTrace()
}

func (w *withStack) Level(level syslog.Level) *withLevel {
	return WithLevel(w, level)
}

func (w *withStack) Status(status int) *withStatus {
	return WithStatus(w, status)
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

func (w *withMessage) Unwrap() error {
	return w.cause
}

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Unwrap())
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

// Unwrap returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type unwrapper interface {
//            Unwrap() error
//     }
//
// If the error does not implement Unwrap, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Unwrap(err error) error {
	for err != nil {
		unwrapper, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = unwrapper.Unwrap()
	}
	return err
}

type unwrapper interface {
	Unwrap() error
}
