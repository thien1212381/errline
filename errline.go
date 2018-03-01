/*
Package adds file and line information to Go errors.

Errors in Go do not have stack trace or file location information attached to them.
Stack traces provide extensive location failure information but are very verbose.
This package only attaches short file name and line information to the errors.

Usage:

1. Do not use it inside generic packages. Your libraries should simply return errors. That's it.
2. Use it in application level packages and application code to capture failure point information.

Example:

// Your application library packages or main package.

func SomeWork() error {
	// Here we call some library function.
	if err := ftp.Connect(addr); err != nil {
		// Here we wrap error with file and line information.
		// If error is already wrapped with file information then Wrap returns original error.
		return errline.Wrap(err);
	}
}

func main() {
	if err := SomeWork; err != nil {
		// IMPORTANT: only +v verb will print short file name and line number.
		// Other verbs simply print err.Error() without file information.
		log.Printf("%+v", err)
	}
}

*/
package errline

import (
	"fmt"
	"io"
	"runtime"
)

const calldepth = 1

// Wrap annotates err with file and line at the point Wrap was first called.
// If err is nil, Wrap returns nil.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	// If error already has file line do not add it again.
	if _, ok := err.(*withFileLine); ok {
		return err
	}

	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}

	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	return &withFileLine{err, file, line}
}

type withFileLine struct {
	error
	file string
	line int
}

func (w *withFileLine) Cause() error { return w.error }

func (w *withFileLine) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s:%d: %+v", w.file, w.line, w.Cause())
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}
