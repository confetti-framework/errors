# errors [![Travis-CI](https://travis-ci.org/confetti-framework/errors.svg)](https://travis-ci.org/confetti-framework/errors) [![AppVeyor](https://ci.appveyor.com/api/projects/status/b98mptawhudj53ep/branch/master?svg=true)](https://ci.appveyor.com/project/davecheney/errors/branch/master) [![GoDoc](https://godoc.org/github.com/confetti-framework/errors?status.svg)](http://godoc.org/github.com/confetti-framework/errors) [![Report card](https://goreportcard.com/badge/github.com/confetti-framework/errors)](https://goreportcard.com/report/github.com/confetti-framework/errors) [![Sourcegraph](https://sourcegraph.com/github.com/confetti-framework/errors/-/badge.svg)](https://sourcegraph.com/github.com/confetti-framework/errors?badge)

Package errors provides simple error handling primitives.

`go get github.com/confetti-framework/errors`

The traditional error handling idiom in Go is roughly akin to
```go
if err != nil {
        return err
}
```
which applied recursively up the call stack results in error reports without context or debugging information. The errors package allows programmers to add context to the failure path in their code in a way that does not destroy the original value of the error.

## Adding context to an error

The errors.Wrap function returns a new error that adds context to the original error. For example
```go
_, err := ioutil.ReadAll(r)
if err != nil {
        return errors.Wrap(err, "read failed")
}
```
## Retrieving the cause of an error

Using `errors.Wrap` constructs a stack of errors, adding context to the preceding error. Depending on the nature of the
error it may be necessary to reverse the operation of errors.Wrap to retrieve the original error for inspection. Any
error value which implements this interface can be inspected by `errors.Unwrap`.
```go
type causer interface {
        Unwrap() error
}
```

`errors.Unwrap` will recursively retrieve the topmost error which does not implement `causer`, which is assumed to be
the original cause. For example:
```go
switch err := errors.Unwrap(err).(type) {
case *MyError:
        // handle specifically
default:
        // unknown error
}
```

[Read the package documentation for more information](https://godoc.org/github.com/confetti-framework/errors).

## Contributing

Because of the Go2 errors changes, this package is not accepting proposals for new functionality. With that said, we welcome pull requests, bug fixes and issue reports. 

Before sending a PR, please discuss your change by raising an issue.

## License

BSD-2-Clause
