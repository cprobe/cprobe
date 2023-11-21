package httpserver

// ErrorWithStatusCode is error with HTTP status code.
//
// The given StatusCode is sent to client when the error is passed to Errorf.
type ErrorWithStatusCode struct {
	Err        error
	StatusCode int
}

// Unwrap returns e.Err.
//
// This is used by standard errors package. See https://golang.org/pkg/errors
func (e *ErrorWithStatusCode) Unwrap() error {
	return e.Err
}

// Error implements error interface.
func (e *ErrorWithStatusCode) Error() string {
	return e.Err.Error()
}
