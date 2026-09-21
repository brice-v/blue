package ml

import "fmt"

// recoverAsError converts a panic raised by the underlying borncgo backend into
// an error. borncgo validates shapes and arguments with panics, and ml is the
// layer that returns errors, so this is where those panics become values blue
// can catch instead of taking the process down.
func recoverAsError(errp *error) {
	if r := recover(); r != nil {
		*errp = fmt.Errorf("%v", r)
	}
}
